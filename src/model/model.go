package model

import (
	"github.com/zgwit/iot-master-next/src/plugin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KeyValueType map[string]interface{}
type KeyValueStringType map[string]string
type KeyValueNumberType map[string]int64
type KeyValueFloatType map[string]float64
type KeyValueBoolType map[string]bool

type TimeValueType struct {
	Time  int64       `form:"time" bson:"time" json:"time"`
	Value interface{} `form:"value" bson:"value" json:"value"`
}

type TimeValueStringType struct {
	Time  int64  `form:"time" bson:"time" json:"time"`
	Value string `form:"value" bson:"value" json:"value"`
}

type TimeValueBoolType struct {
	Time  int64 `form:"time" bson:"time" json:"time"`
	Value bool  `form:"value" bson:"value" json:"value"`
}

type NameIdType struct {
	Name string `form:"name" bson:"name" json:"name"`
	Id   string `form:"id" bson:"id" json:"id"`
}

type PageType struct {
	Start int64 `form:"start" bson:"start" json:"start"`
	Stop  int64 `form:"stop" bson:"stop" json:"stop"`

	Size   int64 `form:"size" bson:"size" json:"size"`
	Offset int64 `form:"offset" bson:"offset" json:"offset"`

	Desc bool `form:"desc" bson:"desc" json:"desc"`
}

// 测试222 - 1
func FetchList(mongo *plugin.Mongo, database string, filter plugin.BSON, tables any) (err error) {

	return mongo.FindAll(database, filter, tables)
}

func FetchList2(mongo *plugin.Mongo, database string, filter plugin.BSON) (tables []map[string]interface{}, err error) {

	if err = mongo.FindAll(database, filter, &tables); err != nil {
		return
	}

	if len(tables) == 0 {
		tables = []map[string]interface{}{}
	}

	for index := range tables {
		if value, exist := tables[index]["_id"]; exist {
			tables[index]["objectid"] = value
			delete(tables[index], "_id")
		}
	}

	return
}

func TableCreate(mongo *plugin.Mongo, database string, table any) (err error) {

	_, err = mongo.InsertOne(database, table)

	return
}

func TableDelete(mongo *plugin.Mongo, database string, objectid primitive.ObjectID) (err error) {

	_, err = mongo.DeleteOne(database, plugin.BSON{"_id": objectid})

	return
}

func TableFind(mongo *plugin.Mongo, database string, filter plugin.BSON, table any) (result bool) {

	return mongo.FindOne(database, filter, table) == nil
}

func TableExist(mongo *plugin.Mongo, database string, filter plugin.BSON) (table map[string]interface{}, result bool) {

	result = mongo.FindOne(database, filter, &table) == nil

	if value, exist := table["_id"]; exist {
		table["objectid"] = value
		delete(table, "_id")
	}

	return
}

func TableUpdate(mongo *plugin.Mongo, database string, objectid primitive.ObjectID, option plugin.BSON) (err error) {

	for key := range option {
		if option[key] == nil {
			delete(option, key)
		}
	}

	_, err = mongo.UpdateOne(database, plugin.BSON{"_id": objectid}, plugin.BSON{"$set": option})

	return
}
