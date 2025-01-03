package game

import (
	"matchgame/logger"
	"time"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"

	// serverSDK "agones.dev/agones/pkg/sdk"
	// "agones.dev/agones/pkg/util/signals"
	sdk "agones.dev/agones/sdks/go"
	log "github.com/sirupsen/logrus"
)

var AgonesSDK *sdk.SDK

func InitAgones() error {
	var err error
	AgonesSDK, err = sdk.NewSDK()
	if err != nil {
		log.Errorf("%s 初始化AgonesSDK錯誤: %v.\n", logger.LOG_Main, err)
		return err
	}
	return nil
}

func SetServerState(state agonesv1.GameServerState) {
	if AgonesSDK == nil {
		log.Errorf("%s 尚未初始化AgonesSDK", logger.LOG_Main)
		return
	}
	switch state {
	case agonesv1.GameServerStateReady:
		// 將此遊戲房伺服器狀態標示為Ready(要標示為ready才會被Agones Allocation服務分配到)
		if err := AgonesSDK.Ready(); err != nil {
			log.Errorf("%s Server狀態標示為Ready錯誤: %v", logger.LOG_Agones, err)
			return
		} else {
			log.Infof("%s Matchgame(AgonesGameServer)標示為Ready", logger.LOG_Agones)
		}
	case agonesv1.GameServerStateAllocated:
		// 將此遊戲房伺服器狀態標示為Allocated, 代表無法再被分配玩家狀態
		if err := AgonesSDK.Allocate(); err != nil {
			log.Errorf("%s Server狀態標示為Allocated錯誤: %v", logger.LOG_Agones, err)
			return
		} else {
			log.Infof("%s Matchgame(AgonesGameServer)標示為Allocated", logger.LOG_Agones)
		}
	}

}

// 通知Agones關閉server，pod會被回收
func ShutdownAgonesServer() {
	if AgonesSDK == nil {
		log.Errorf("%s 尚未初始化AgonesSDK", logger.LOG_Main)
		return
	}
	// 通知Agones關閉server
	if err := AgonesSDK.Shutdown(); err != nil {
		log.Errorf("%s 通知Agones關閉server錯誤: %v", logger.LOG_Main, err)
	}
}

// 送定時送Agones健康ping通知agones server遊戲房還活著
// Agones的超時為periodSeconds設定的秒數 參考官方: https://agones.dev/site/docs/guides/health-checking/
func AgonesHealthPin(stop <-chan struct{}) {
	if AgonesSDK == nil {
		log.Errorf("%s 尚未初始化AgonesSDK", logger.LOG_Main)
		return
	}
	tick := time.Tick(AGONES_HEALTH_PIN_INTERVAL_SEC * time.Second)
	for {
		if err := AgonesSDK.Health(); err != nil {
			log.Errorf("%s ping agones server錯誤: %v", logger.LOG_Main, err)
		}
		select {
		case <-stop:
			log.Infof("%s Health pings 意外停止", logger.LOG_Main)
			return
		case <-tick:
		}
	}
}

// 延遲關閉Agones server
func DelayShutdownServer(delay time.Duration, stop chan struct{}) {
	timer1 := time.NewTimer(delay)
	<-timer1.C
	// 通知Agones關閉server
	if err := AgonesSDK.Shutdown(); err != nil {
		log.Errorf("%s Could not call shutdown: %v", logger.LOG_Main, err)
	}
	stop <- struct{}{}
}
