package game

import (
	"context"
	"fmt"
	"lobby/logger"
	"time"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	allocationv1 "agones.dev/agones/pkg/apis/allocation/v1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	gameserverName       = "gladiators-matchgame"
	gameserversNamespace = "gladiators-gameserver"
)

var agonesClient *versioned.Clientset

// 初始化Agones Client
func InitAgones() error {
	log.Infof("%s 開始初始化Agones API Client", logger.LOG_Agones)
	var err error
	// 取目前pod所在k8s cluster的config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("%s 取cluster的config錯誤: %v", logger.LOG_Agones, err)
		return err
	}
	// 與agones連接
	agonesClient, err = versioned.NewForConfig(config)
	if err != nil {
		log.Errorf("%s 建立 agones api client錯誤: %v", logger.LOG_Agones, err)
		return err
	}
	log.Infof("%s 初始化Agones API Client完成", logger.LOG_Agones)
	return nil
}

// GetAgonesClient 取得Agones API Client
func GetAgonesClient() *versioned.Clientset {
	if agonesClient == nil {
		log.Infof("%s 尚未建立 agones api client, 嘗試建立", logger.LOG_Agones)
		err := InitAgones()
		if err != nil {
			return nil
		}
	}
	return agonesClient
}

// createGameServerAllocation 建立GameServerAllocation配置
func createGameServerAllocation(lobbyPodName string, labels map[string]string) *allocationv1.GameServerAllocation {
	return &allocationv1.GameServerAllocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s_allocation", lobbyPodName),
			Namespace: gameserversNamespace,
		},
		Spec: allocationv1.GameServerAllocationSpec{
			Required: allocationv1.GameServerSelector{
				LabelSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{"agones.dev/fleet": gameserverName},
				},
			},
			MetaPatch: allocationv1.MetaPatch{
				Labels: labels,
			},
		},
	}
}

// allocateGameServer 分配GameServer並返回結果
func allocateGameServer(allocation *allocationv1.GameServerAllocation) (*agonesv1.GameServer, error) {
	allocateInterface := agonesClient.AllocationV1().GameServerAllocations(gameserversNamespace)
	gsAllocation, err := allocateInterface.Create(context.Background(), allocation, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("分配 game server 失敗: %v", err)
	}

	newGS, err := agonesClient.AgonesV1().GameServers(gameserversNamespace).Get(
		context.Background(),
		gsAllocation.Status.GameServerName,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("取得已分配的 game server 失敗: %v", err)
	}

	return newGS, nil
}

// ApplyGameServer 分配遊戲房伺服器
func ApplyGameServer(roomID string, playerIDs []string, createrID string, dbMapID string, lobbyPodName string) (*agonesv1.GameServer, error) {
	// 建立通用的標籤
	labels := map[string]string{
		"RoomName":     roomID,
		"CreaterID":    createrID,
		"LobbyPodName": lobbyPodName,
		"DBMapID":      dbMapID,
	}

	for i, playerID := range playerIDs {
		labels[fmt.Sprintf("Player%d", i)] = playerID
	}

	// 先嘗試分配現有的空閒GameServer
	if gs, err := tryAllocateExistingServer(labels, lobbyPodName); err == nil {
		return gs, nil
	}

	// 如果沒有可用的，建立新的GameServer
	allocation := createGameServerAllocation(lobbyPodName, labels)
	newGS, err := allocateGameServer(allocation)
	if err != nil {
		return nil, err
	}

	// 等待 GameServer 準備就緒
	if err := waitForGameServerReady(newGS.ObjectMeta.Name); err != nil {
		return nil, fmt.Errorf("等待 GameServer 準備就緒失敗: %v", err)
	}

	log.Infof("%s 建立新的遊戲房間: %s, address=%s, port=%v",
		logger.LOG_Agones,
		newGS.ObjectMeta.Name,
		newGS.Status.Address,
		newGS.Status.Ports[0].Port)

	return newGS, nil
}

// waitForGameServerReady 等待 GameServer 準備就緒
func waitForGameServerReady(gsName string) error {
	// 最多等待 GAMESERVER_WAIT_TIME 秒
	for i := 0; i < GAMESERVER_WAIT_TIME; i++ {
		gs, err := agonesClient.AgonesV1().GameServers(gameserversNamespace).Get(
			context.Background(),
			gsName,
			metav1.GetOptions{},
		)
		if err != nil {
			return err
		}

		// 檢查 GameServer 是否準備就緒
		if gs.Status.State == agonesv1.GameServerStateReady {
			return nil
		}
		// 等待 1 秒後再次檢查
		time.Sleep(time.Second)
	}

	return fmt.Errorf("等待 GameServer 準備就緒超時")
}

// tryAllocateExistingServer 嘗試分配現有的空閒GameServer
func tryAllocateExistingServer(labels map[string]string, lobbyPodName string) (*agonesv1.GameServer, error) {
	gameServers, err := agonesClient.AgonesV1().GameServers(gameserversNamespace).List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: "agones.dev/fleet=" + gameserverName,
		})
	if err != nil {
		return nil, fmt.Errorf("查詢可用 game server 失敗: %v", err)
	}

	for _, gs := range gameServers.Items {
		if gs.Status.State == agonesv1.GameServerStateReady {
			allocation := createGameServerAllocation(lobbyPodName, labels)
			if newGS, err := allocateGameServer(allocation); err == nil {
				log.Infof("%s 成功分配現有 game server: %s", logger.LOG_Agones, newGS.ObjectMeta.Name)
				return newGS, nil
			}
		}
	}
	return nil, fmt.Errorf("no available game server found")
}

// 檢查該Matchgame server是正常運作
func CheckGameServer(roomID string) error {
	gameServers, err := agonesClient.AgonesV1().GameServers(gameserversNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Errorf("取得Matchgame列表失敗: %v", err)
		return err
	}
	log.Infof("檢查RoomName是否可加入: %s", roomID)
	for _, gs := range gameServers.Items {
		log.Infof("GS RoomName: %s", gs.Labels["RoomName"])
		if gs.Labels["RoomName"] == roomID {
			log.Infof("目標RoomName狀態: %s", gs.Status.State)
			if gs.Status.State == agonesv1.GameServerStateAllocated {
				log.Infof("%s 已確認目標Matchgame server正常運作", logger.LOG_Agones)
				return nil
			} else {
				return fmt.Errorf("%s Matchgame(%s)掛了", logger.LOG_Agones, roomID)
			}
		}
	}
	return fmt.Errorf("%s Matchgame(%s)不存在", logger.LOG_Agones, roomID)
}
