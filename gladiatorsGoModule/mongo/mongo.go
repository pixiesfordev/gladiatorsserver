package mongo

import (
	"context"
	"fmt"
	"time"

	"gladiatorsGoModule/logger"
	"gladiatorsGoModule/setting"

	log "github.com/sirupsen/logrus"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	db     *mongoDriver.Database
	client *mongoDriver.Client

	// 初始化資料 (Singleton)
	self InitData = InitData{
		Env:           "Dev", // 目前的環境版本
		APIPublicKey:  "",    // 目前的 Realm 的 APIKey
		APIPrivateKey: "",    // 目前的 Realm 的 APIKey
	}
)

type InitData struct {
	Env           string
	APIPublicKey  string
	APIPrivateKey string
}

// Singleton
func DB() *mongoDriver.Database {
	return db
}

func Init(data InitData, user string, pw string) {
	self = data

	connect(user, pw) // 連線 MongoDB
}

// 連線 MongoDB
func connect(user string, pw string) error {
	if client != nil {
		return fmt.Errorf("MongoDB client 已經被建立，不可重複建立")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	uri := setting.MongoURI(self.Env, user, pw)
	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))
	newClient, err := mongoDriver.Connect(ctx, clientOptions)
	if err != nil {
		log.Errorf("%s 連線 MongoDB 失敗: %v", logger.LOG_Mongo, err)
		return err
	}

	db = newClient.Database(setting.MongoDBName(self.Env))

	return nil
}

// 斷開與 MongoDB 的連線
func Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if client != nil {
		client.Disconnect(ctx)
	}
}
