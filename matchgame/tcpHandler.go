package main

import (
	"crypto/rand"
	logger "matchgame/logger"
	gSetting "matchgame/setting"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"encoding/hex"
	"encoding/json"
	mongo "gladiatorsGoModule/mongo"
	"gladiatorsGoModule/redis"
	"matchgame/game"
	"matchgame/packet"
	"net"
	"time"

	sdk "agones.dev/agones/sdks/go"
)

// 開啟TCP連線
func openConnectTCP(s *sdk.SDK, stop chan struct{}, src string) {
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
		if err := recover(); err != nil {
			log.Errorf("%s (TCP)handleConnectionTCP錯誤: %v.", logger.LOG_Main, err)
		}
	}()
	isAuth := false
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
	for {
		select {
		case <-stop:
			log.Errorf("強制終止TCP")
			// 被強制終止
			return
		default:
		}
		pack, err := packet.ReadPack(decoder)
		if err != nil {
			game.MyRoom.KickPlayer(conn, "handleConnectionTCP錯誤")
			return
		}
		log.Infof("%s (TCP)收到來自 %s 的命令: %s \n", logger.LOG_Main, remoteAddr, pack.CMD)

		//未驗證前，除了Auth指令進來其他都擋掉
		if !isAuth && pack.CMD != packet.AUTH {
			log.Infof("%s 收到未驗證的封包", logger.LOG_Main)
			return
		}
		if pack.CMD == packet.AUTH {
			authContent := packet.Auth{}
			if ok := authContent.Parse(pack.Content); !ok {
				log.Errorf("%s 反序列化AUTH封包失敗", logger.LOG_Main)
				return
			}
			// 像mongodb atlas驗證token並取得playerID 有通過驗證後才處理後續
			playerID, authErr := mongo.PlayerVerify(authContent.Token)
			// 驗證失敗
			if authErr != nil || playerID == "" {
				log.Errorf("%s 玩家驗證錯誤: %v", logger.LOG_Main, authErr)
				_ = packet.SendPack(encoder, &packet.Pack{
					CMD:    packet.AUTH_TOCLIENT,
					PackID: pack.PackID,
					ErrMsg: "玩家驗證錯誤",
					Content: &packet.Auth_ToClient{
						IsAuth: false,
					},
				})
			}
			var dbPlayer mongo.DBPlayer
			getPlayerDocErr := mongo.GetDocByID(mongo.ColName.Player, playerID, &dbPlayer)
			if getPlayerDocErr != nil {
				log.Errorf("%s DBPlayer資料錯誤: %v", logger.LOG_Main, getPlayerDocErr)
				_ = packet.SendPack(encoder, &packet.Pack{
					CMD:    packet.AUTH_TOCLIENT,
					PackID: pack.PackID,
					ErrMsg: "DBPlayer資料錯誤",
					Content: &packet.Auth_ToClient{
						IsAuth: false,
					},
				})
			}

			isAuth = true
			// 建立RedisDB Player
			redisPlayer, redisPlayerErr := redis.CreatePlayerData(dbPlayer.ID, int(dbPlayer.Point), int(dbPlayer.HeroExp), dbPlayer.SpellCharges, dbPlayer.Drops)
			if redisPlayerErr != nil {
				log.Errorf("%s 建立RedisPlayer錯誤: %v", logger.LOG_Main, redisPlayerErr)
				_ = packet.SendPack(encoder, &packet.Pack{
					CMD:    packet.AUTH_TOCLIENT,
					PackID: pack.PackID,
					ErrMsg: "建立RedisPlayer錯誤",
					Content: &packet.Auth_ToClient{
						IsAuth: false,
					},
				})
			}
			redisPlayer.StartInGameUpdatePlayer() // 開始跑玩家資料定時更新上RedisDB程序

			// 將該玩家monogoDB上的redisSync設為false
			updatePlayerBson := bson.D{
				{Key: "redisSync", Value: false},
			}
			mongo.UpdateDocByBsonD(mongo.ColName.Player, dbPlayer.ID, updatePlayerBson)

			// 建立udp socket連線Token
			newConnToken := generateSecureToken(32)

			// 將玩家加入遊戲房
			player := game.Player{
				DBPlayer:     &dbPlayer,
				RedisPlayer:  redisPlayer,
				LastUpdateAt: time.Now(),
				PlayerBuffs:  []packet.PlayerBuff{},
				ConnTCP: &gSetting.ConnectionTCP{
					Conn:    conn,
					Encoder: encoder,
					Decoder: decoder,
				},
				ConnUDP: &gSetting.ConnectionUDP{
					ConnToken: newConnToken,
				},
			}
			joined := game.MyRoom.JoinPlayer(&player)
			if !joined {
				log.Errorf("%s 玩家加入房間失敗", logger.LOG_Main)
				return
			}
			// 回送client
			err = packet.SendPack(encoder, &packet.Pack{
				CMD:    packet.AUTH_TOCLIENT,
				PackID: pack.PackID,
				Content: &packet.Auth_ToClient{
					IsAuth:    true,
					ConnToken: newConnToken,
					Index:     player.Index,
				},
			})
			if err != nil {
				return
			}

		} else {
			err = game.MyRoom.HandleTCPMsg(conn, pack)
			if err != nil {
				log.Errorf("%s (TCP)處理GameRoom封包錯誤: %v\n", logger.LOG_Main, err.Error())
				game.MyRoom.KickPlayer(conn, "處理GameRoom封包錯誤")
				return
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
