package main

import (
	"context"
	"encoding/json"
	"gladiatorsGoModule/mongo"
	"lobby/game"
	"lobby/logger"
	"lobby/packet"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

func initTcp(port string) {
	// 偵聽TCP封包
	src := ":" + port
	tcpListener, err := net.Listen("tcp", src)
	if err != nil {
		log.Errorf("%s  (TCP)偵聽錯誤 %s.\n", logger.LOG_TCP, err.Error())
		return
	}
	defer tcpListener.Close()
	log.Infof("%s (TCP)開始偵聽: %s", logger.LOG_TCP, src)

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			log.Errorf("%s (TCP)x連線錯誤 %s.\n", logger.LOG_TCP, err)
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		go handleConnectionTCP(ctx, conn, cancel)
	}
}

// 處理TCP連線封包
func handleConnectionTCP(ctx context.Context, conn net.Conn, cancel context.CancelFunc) {
	remoteAddr := conn.RemoteAddr().String()
	log.Infof("%s 連線來自IP: %s", logger.LOG_TCP, remoteAddr)

	defer func() {
		log.Infof("%s (TCP)關閉%v的連線", logger.LOG_TCP, remoteAddr)
		conn.Close()
		cancel()
	}()
	conn.SetReadDeadline(time.Now().Add(game.CONN_TIMEOUT_SECONDS * time.Second)) // 設定超時
	decoder := json.NewDecoder(conn)
	isAuth := false
	var player *game.Player
	// 封包處理
	for {
		select {
		case <-ctx.Done(): // 檢查上下文是否被取消
			log.Errorf("%s (TCP)強制終止%v的連線", logger.LOG_TCP, remoteAddr)
			return
		default:
			pack, err := packet.ReadPack(decoder)
			if err != nil {
				log.Errorf("%s (TCP)讀取封包錯誤: %v.", logger.LOG_TCP, err)
				return
			}
			conn.SetReadDeadline(time.Now().Add(game.CONN_TIMEOUT_SECONDS * time.Second))
			if pack.CMD == packet.AUTH {
				content := packet.Auth{}
				err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
				if err != nil {
					log.Errorf("%s (TCP)Auth解包錯誤: %v.", logger.LOG_TCP, err)
					return
				}
				encoder := json.NewEncoder(conn)
				dbPlayer, authErr := mongo.VerifyPlayerByToken(content.ConnToken)
				if authErr != nil || dbPlayer == nil {
					log.Errorf("%s %v驗證錯誤: %v", logger.LOG_TCP, remoteAddr, authErr)
					packet.SendPack(encoder, packet.Pack{
						CMD:    packet.AUTH_TOCLIENT,
						PackID: pack.PackID,
						ErrMsg: "驗證錯誤",
						Content: &packet.Auth_ToClient{
							IsAuth: false,
							Time:   time.Now().UnixMilli(),
						},
					})
					cancel()
					return
				}
				player, err = game.NewPlayer(dbPlayer.ID, conn)
				if err != nil {
					log.Errorf("%s (TCP)建立玩家錯誤: %v", logger.LOG_TCP, err)
					cancel()
					return
				}
				isAuth = true
				// 回送client
				err = packet.SendPack(encoder, packet.Pack{
					CMD:    packet.AUTH_TOCLIENT,
					PackID: pack.PackID,
					Content: packet.Auth_ToClient{
						IsAuth: isAuth,
					},
				})
				if err != nil {
					log.Errorf("%s (TCP)回送Auth_ToClient錯誤: %v", logger.LOG_TCP, err)
					cancel()
					return
				}

			} else {
				if !isAuth {
					log.Errorf("%s (TCP)收到來自 %v 的未驗證封包: %v", logger.LOG_TCP, remoteAddr, pack.CMD)
					cancel()
					return
				}
				if pack.CMD != packet.PING {
					log.Infof("%s (TCP)收到來自%s 的命令: %s \n", logger.LOG_TCP, remoteAddr, pack.CMD)
				}
				err := game.HandleTCPMsg(player, pack)
				if err != nil {
					log.Errorf("%s (TCP)HandleTCPMsg錯誤: %v", logger.LOG_TCP, err)
					cancel()
					return
				}
			}
		}
	}
}
