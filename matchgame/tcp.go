package main

import (
	"crypto/rand"
	"fmt"
	logger "matchgame/logger"
	"sync"

	log "github.com/sirupsen/logrus"

	"encoding/hex"
	"encoding/json"
	mongo "gladiatorsGoModule/mongo"
	"matchgame/game"
	"matchgame/packet"
	"net"
	"time"
	// sdk "agones.dev/agones/sdks/go"
)

// 開啟TCP連線
func openConnectTCP(stop chan struct{}, src string) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s OpenConnectTCP error: %v.\n", logger.LOG_Main, err)
			stop <- struct{}{}
		}
	}()
	tcpListener, err := net.Listen("tcp", src)
	if err != nil {
		log.Errorf("%s (TCP)偵聽失敗: %v.\n", logger.LOG_Main, err)
	}
	defer tcpListener.Close()
	log.Infof("%s (TCP)開始偵聽 %s", logger.LOG_Main, src)
	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			log.Errorf("%s Unable to accept incoming tcp connection: %v.\n", logger.LOG_Main, err)
			continue
		}
		go handleConnectionTCP(conn, stop)
	}
}

// 處理TCP連線封包
func handleConnectionTCP(conn net.Conn, stop chan struct{}) {
	remoteAddr := conn.RemoteAddr().String()

	// log.Infof("%s Client %s connected", logger.LOG_Main, conn.RemoteAddr().String())
	defer conn.Close()
	defer func() {
		log.Infof("%s 關閉handleConnectionTCP", logger.LOG_Main)
		// if err := recover(); err != nil {
		// 	// log.Errorf("%s (TCP)handleConnectionTCP錯誤: %v.", logger.LOG_Main, err)
		// }
	}()
	isAuth := false
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	conn.SetReadDeadline(time.Now().Add(game.TCP_CONN_TIMEOUT_SEC * time.Second))

	packReadChan := &game.MyChan{
		StopChan:      make(chan struct{}, 1),
		ChanCloseOnce: sync.Once{},
	}

	// 封包處理
	for {
		select {
		case <-stop:
			log.Errorf("%s (TCP)強制終止連線", logger.LOG_Main)
			packReadChan.Close()
			return
		case <-packReadChan.StopChan:
			log.Infof("%s (TCP)關閉封包讀取", logger.LOG_Main)
			return // 終止goroutine
		default:
			pack, err := packet.ReadPack(decoder)
			if err != nil {
				log.Infof("%s (TCP)packReadReadResult錯誤: %v.", logger.LOG_Main, err)
				packReadChan.Close()
				return
			}
			if pack.CMD != packet.PING {
				log.Infof("%s (TCP)收到來自%s 的命令: %s \n", logger.LOG_Main, remoteAddr, pack.CMD)
			}

			//未驗證前，除了Auth指令進來其他都擋掉
			if !isAuth && pack.CMD != packet.AUTH {
				log.Infof("%s 收到未驗證的封包", logger.LOG_Main)
				packReadChan.Close()
				return
			}
			if pack.CMD == packet.AUTH {
				authContent := packet.Auth{}
				err := json.Unmarshal([]byte(pack.GetContentStr()), &authContent)
				if err != nil {
					log.Errorf("%s (TCP)Auth解包錯誤: %v.", logger.LOG_Main, err)
					return
				}
				// 像mongodb atlas驗證token並取得playerID 有通過驗證後才處理後續
				dbPlayer, authErr := mongo.VerifyPlayerByToken(authContent.ConnToken)
				if authErr != nil || dbPlayer == nil {
					errLog := fmt.Sprintf("%s mongo.VerifyPlayerByToken錯誤, Token: %v", logger.LOG_Main, authContent.ConnToken)
					_ = packet.SendPack(encoder, packet.Pack{
						CMD:    packet.AUTH_TOCLIENT,
						PackID: pack.PackID,
						ErrMsg: errLog,
						Content: packet.Auth_ToClient{
							IsAuth: false,
						},
					})
					continue
				}
				isAuth = true
				var player *game.Player
				// 斷線重連檢測
				reConnection := false
				for _, v := range game.MyRoom.Gamers {
					if p, ok := v.(*game.Player); ok {
						if p.GetID() == dbPlayer.ID {
							log.Infof("玩家(%v)斷線重連", dbPlayer.ID)
							reConnection = true
							player = p
							break
						}
					}
				}

				// 建立udp socket連線Token
				newConnToken := generateSecureToken(32)

				if !reConnection { // 不是斷線重連就建立玩家資料
					dbPlayer, getPlayerDocErr := mongo.GetDocByID[mongo.DBPlayer](mongo.Col.Player, dbPlayer.ID)
					if getPlayerDocErr != nil {
						log.Errorf("%s DBPlayer資料錯誤: %v", logger.LOG_Main, getPlayerDocErr)
						_ = packet.SendPack(encoder, packet.Pack{
							CMD:    packet.AUTH_TOCLIENT,
							PackID: pack.PackID,
							ErrMsg: "DBPlayer資料錯誤",
							Content: &packet.Auth_ToClient{
								IsAuth: false,
							},
						})
					}
					connTCP := &game.ConnectionTCP{
						Conn:       conn,
						MyLoopChan: packReadChan,
						Encoder:    encoder,
						Decoder:    decoder,
					}
					connUDP := &game.ConnectionUDP{
						ConnToken: newConnToken,
					}
					// 將玩家加入遊戲房
					player = game.NewPlayer(dbPlayer.ID, connTCP, connUDP)
					log.Infof("player: %v ", player)
					err = game.MyRoom.JoinGamer(player)
					if err != nil {
						log.Errorf("%s 玩家加入房間失敗: %v", logger.LOG_Main, err)
						packReadChan.Close()
						return
					}
				} else { // 斷線重連時使用已存在的玩家資料, 不須重建資料
					player.ConnTCP.Conn = conn
					player.ConnUDP.ConnToken = newConnToken
				}
				// 回送client
				err = packet.SendPack(encoder, packet.Pack{
					CMD:    packet.AUTH_TOCLIENT,
					PackID: pack.PackID,
					Content: packet.Auth_ToClient{
						IsAuth: true,
					},
				})
				if err != nil {
					continue
				}
			} else {
				err = game.HandleTCPMsg(conn, pack)
				if err != nil {
					log.Errorf("%s (TCP)處理GameRoom封包錯誤: %v\n", logger.LOG_Main, err.Error())
					// player := game.MyRoom.GetPlayerByTCPConn(conn)
					// if player != nil {
					// 	game.MyRoom.KickPlayer(player, "處理GameRoom封包錯誤")
					// }
					continue
				}
			}
		}

	}
}

// 產生連線驗證Token
func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
