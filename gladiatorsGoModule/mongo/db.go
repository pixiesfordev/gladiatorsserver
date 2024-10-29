package mongo

// MongoDB go CRUD參考官方文件: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/write-operations/modify/

import (
	"context"
	"errors"

	"gladiatorsGoModule/logger"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	errDBNotInitialized = errors.New("database is not initialized")
	errColNotExist      = errors.New("collection does not exist")
)

// 使用 Collection 與 ID 來取文件
func GetDocByID(col string, id string, result interface{}) error {
	if db == nil {
		return errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return errColNotExist
	}

	err := collection.FindOne(context.TODO(),
		bson.D{{Key: "_id", Value: id}}, // filter
	).Decode(result)
	if err != nil {
		log.Infof("%s GetDocByID 錯誤: %v", logger.LOG_Mongo, err)
		return err
	}
	return nil
}

// 使用 Collection 與 Filter 來取第一個符合條件的文件
func GetDocByFilter[T any](col string, filter bson.M) (*T, error) {
	if db == nil {
		return nil, errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return nil, errColNotExist
	}

	var result bson.M
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Infof("%s GetDocByFilter 錯誤: %v", logger.LOG_Mongo, err)
		return nil, err
	}

	return BSONToStruct[T](result)
}

// 使用 Collection 與 Filter 來取文件
func GetDocsByFilter(col string, filter bson.M, results interface{}) error {
	if db == nil {
		return errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return errColNotExist
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return err
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), results); err != nil {
		return err
	}

	return nil
}

// GetDocsByFieldValue 取得指定欄位符合特定條件下的所有文件
// fieldValue 必須符合 fieldName 的類型
// results 必須是 DB Sechma Struct 切片 Ex: []DBPlayer
func GetDocsByFieldValue(col string, fieldName string, fieldValue interface{}, operator Operator, results interface{}) error {
	if db == nil {
		return errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return errColNotExist
	}

	cursor, err := collection.Find(context.TODO(),
		bson.M{fieldName: bson.M{string(operator): fieldValue}}, // filterName : filterValue
	)
	if err != nil {
		return err
	}
	defer cursor.Close(context.TODO())

	// 解碼查詢結果到 results 指向的切片
	if err = cursor.All(context.TODO(), results); err != nil {
		return err
	}

	return nil
}

// GetDocsByFieldValue 取得指定欄位符合特定條件下的所有文件的 ID slice
// fieldValue 必須符合 fieldName 的類型
func GetDocIDsByFieldValue(col string, fieldName string, fieldValue interface{}, operator Operator) ([]string, error) {
	if db == nil {
		return nil, errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return nil, errColNotExist
	}

	// 執行查詢只返回 _id
	var results []string
	cursor, err := collection.Find(context.TODO(),
		bson.M{fieldName: bson.M{string(operator): fieldValue}}, // fieldName : fieldValue
		options.Find().SetProjection(bson.M{"_id": 1}),          // 只包括 _id 字段的投影，1 (包含)，0 (排除)
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var elem struct {
			ID string `bson:"_id"`
		}
		if err = cursor.Decode(&elem); err != nil {
			return nil, err
		}
		results = append(results, elem.ID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetDocIDsByFilter 取得符合特定過濾條件的所有文檔的 ID slice
// col: 集合名稱
// filter: 過濾條件
func GetDocIDsByFilter(col string, filter bson.M) ([]string, error) {
	if db == nil {
		return nil, errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return nil, errColNotExist
	}

	// 執行查詢只返回符合投影條件的字段
	var results []string
	cursor, err := collection.Find(context.TODO(), filter,
		options.Find().SetProjection(bson.M{"_id": 1}), // 只包括 _id 字段的投影，1 (包含)，0 (排除)
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var elem struct {
			ID string `bson:"_id"`
		}
		if err = cursor.Decode(&elem); err != nil {
			return nil, err
		}
		results = append(results, elem.ID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// 更新文件
func UpdateDocByBsonD(col string, id string, updateData bson.D) (*mongoDriver.UpdateResult, error) {
	if db == nil {
		return nil, errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return nil, errColNotExist
	}

	result, err := collection.UpdateOne(context.TODO(),
		bson.D{{Key: "_id", Value: id}},          // filter
		bson.D{{Key: "$set", Value: updateData}}, // update
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateDocByStruct 使用struct更新文件
func UpdateDocByStruct(col string, id string, updateData interface{}) (*mongoDriver.UpdateResult, error) {
	if db == nil {
		return nil, errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return nil, errColNotExist
	}

	// 根據 _id 過濾條件找到需要更新的文件
	filter := bson.M{"_id": id}

	// 使用 $set 操作符將 updateData 的內容應用到文件上
	update := bson.M{"$set": updateData}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateDocsByField 批量更新指定欄位符合特定條件的所有文檔
// col: 要更新的集合名
// filterField: 過濾條件的欄位名
// filterValue: 過濾條件的欄位值
// updateData: 更新內容 (bson.D 格式) Ex: bson.D{{Key: "onlineState", Value: "Offline"}}
func UpdateDocsByField(col string, filterField string, filterValue interface{}, updateData bson.D) (*mongoDriver.UpdateResult, error) {
	if db == nil {
		return nil, errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return nil, errColNotExist
	}

	result, err := collection.UpdateMany(context.TODO(),
		bson.M{filterField: bson.M{"$in": filterValue}}, // filterField : filterValue
		bson.D{{Key: "$set", Value: updateData}})        // updateData
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 新增文件
func AddDocByBsonD(col string, addData bson.D) (*mongoDriver.InsertOneResult, error) {
	if db == nil {
		return nil, errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return nil, errColNotExist
	}

	result, err := collection.InsertOne(context.TODO(), addData)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// 新增文件
func AddDocByStruct(col string, addData interface{}) (*mongoDriver.InsertOneResult, error) {
	if db == nil {
		return nil, errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return nil, errColNotExist
	}

	result, err := collection.InsertOne(context.TODO(), addData)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 文件存在就更新不存在就新增
func UpsertDocByStruct(col string, docID string, addData interface{}) (*mongoDriver.UpdateResult, error) {
	if db == nil {
		return nil, errDBNotInitialized
	}

	collection := db.Collection(col)
	if collection == nil {
		return nil, errColNotExist
	}

	result, err := collection.UpdateOne(context.TODO(),
		bson.D{{Key: "_id", Value: docID}},    // filter
		bson.D{{Key: "$set", Value: addData}}, // update
		options.Update().SetUpsert(true),      // options
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}
