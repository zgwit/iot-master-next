package zgwit_plugin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	_mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type BSON bson.M

type Mongo struct {
	Client   *_mongo.Client   `form:"client" bson:"client" json:"client"`
	Database *_mongo.Database `form:"database" bson:"database" json:"database"`
	Config   MongoConfig      `form:"config" bson:"config" json:"config"`
}

type MongoConfig struct {
	Url      string `form:"url" bson:"url" json:"url"`
	Database string `form:"database" bson:"database" json:"database"`
	Username string `form:"username" bson:"username" json:"username"`
	Password string `form:"password" bson:"password" json:"password"`
}

func NewMongo(config MongoConfig) (mongo Mongo) {

	mongo.Config = config
	return
}

func NewShortID() string {

	timer := time.Now().UnixMilli()
	return fmt.Sprintf("%x%x", timer/1000, timer%1000)
}

func NewNullID() (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex("000000000000000000000000")
}

func NewShortObjectID() (primitive.ObjectID, error) {

	timer := time.Now().UnixMilli()
	return primitive.ObjectIDFromHex(fmt.Sprintf("%x%x0000000000000", timer/1000, timer%1000))
}

func (mongo *Mongo) Connect() (err error) {

	if mongo.Client, err = _mongo.Connect(
		context.Background(),
		options.Client().ApplyURI(
			fmt.Sprintf("mongodb://%s:%s@%s/%s", mongo.Config.Username, mongo.Config.Password, mongo.Config.Url, mongo.Config.Database),
		),
	); err == nil {
		mongo.Database = mongo.Client.Database(mongo.Config.Database)
	}

	return
}

func (mongo *Mongo) Disconnect() (err error) {

	return mongo.Client.Disconnect(context.Background())
}

func (mongo *Mongo) Ping(timeOut int) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeOut)*time.Second)
	defer cancel()

	return mongo.Client.Ping(ctx, readpref.Primary())
}

/* --- common --- */

func (mongo *Mongo) FetchList(database string, filter BSON) (tables []map[string]interface{}, err error) {

	if err = mongo.FindAll(database, filter, &tables); err != nil {
		return
	}

	if len(tables) == 0 {
		tables = []map[string]interface{}{}
	}

	return
}

func (mongo *Mongo) TableCreate(database string, table any) (err error) {

	_, err = mongo.InsertOne(database, table)

	return
}

func (mongo *Mongo) TableDelete(database string, objectid primitive.ObjectID) (err error) {

	_, err = mongo.DeleteOne(database, BSON{"_id": objectid})

	return
}

func (mongo *Mongo) TableExist(database string, filter BSON) (table map[string]interface{}, result bool) {

	result = mongo.FindOne(database, filter, &table) == nil

	return
}

func (mongo *Mongo) TableFind(database string, filter BSON, table any) (result bool) {

	return mongo.FindOne(database, filter, table) == nil
}

/* --- base --- */

func (mongo *Mongo) InsertOne(collection string, table interface{}) (result *_mongo.InsertOneResult, err error) {

	result, err = mongo.Database.Collection(collection).InsertOne(
		context.Background(),
		table,
	)

	return
}

func (mongo *Mongo) InsertMany(collection string, tables []map[string]interface{}) (result *_mongo.InsertManyResult, err error) {

	_tables := []interface{}{}

	for idx := range tables {
		_tables = append(_tables, tables[idx])
	}

	result, err = mongo.Database.Collection(collection).InsertMany(
		context.Background(),
		_tables,
	)

	return
}

func (mongo *Mongo) FindOne(collection string, filter map[string]interface{}, tables interface{}) (err error) {

	err = mongo.Database.Collection(collection).FindOne(
		context.Background(),
		filter,
	).Decode(tables)

	return
}

func (mongo *Mongo) FindMany(collection string, filter map[string]interface{}, tables interface{}, index int64, size int64, sort string, order int) (err error) {

	var (
		cursor      *_mongo.Cursor
		findoptions options.FindOptions
	)

	if size > 0 {
		findoptions.SetLimit(size)
		findoptions.SetSkip((index - 1) * size)
	}

	if sort != "" {
		findoptions.SetSort(map[string]interface{}{sort: order})
	}

	if cursor, err = mongo.Database.Collection(collection).Find(context.Background(), filter, &findoptions); err != nil {
		return
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), tables)

	return
}

func (mongo *Mongo) FindAll(collection string, filter map[string]interface{}, tables interface{}) (err error) {

	var (
		cursor      *_mongo.Cursor
		findoptions options.FindOptions
	)

	findoptions.SetSort(map[string]interface{}{"_id": -1})

	if cursor, err = mongo.Database.Collection(collection).Find(context.Background(), filter, &findoptions); err != nil {
		return
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), tables)

	return
}

func (mongo *Mongo) FindAll2(collection string, filter map[string]interface{}, tables interface{}) (err error) {

	var (
		cursor      *_mongo.Cursor
		findoptions options.FindOptions
	)

	if cursor, err = mongo.Database.Collection(collection).Find(context.Background(), filter, &findoptions); err != nil {
		return
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), tables)

	return
}

func (mongo *Mongo) UpdateOne(collection string, filter interface{}, update interface{}) (result *_mongo.UpdateResult, err error) {

	result, err = mongo.Database.Collection(collection).UpdateOne(
		context.Background(),
		filter,
		update,
	)

	return
}

func (mongo *Mongo) ResetOne(collection string, filter interface{}, update interface{}) (result *_mongo.UpdateResult, err error) {

	tableOld := map[string]interface{}{}
	tableNew := map[string]interface{}{}
	delete := map[string]interface{}{}

	err = mongo.Database.Collection(collection).FindOne(
		context.Background(),
		filter,
	).Decode(&tableOld)
	if err != nil {
		err = errors.New("文档不存在")
		return
	}

	for idx := range tableOld {
		exist := false
		for _idx := range update.(map[string]interface{}) {
			if idx == _idx {
				exist = true
				break
			}
		}
		if !exist && idx != "_id" {
			delete[idx] = 1
		}
	}

	tableNew["$set"] = update
	if len(delete) > 0 {
		tableNew["$unset"] = delete
	}

	result, err = mongo.Database.Collection(collection).UpdateOne(
		context.Background(),
		filter,
		tableNew,
	)

	return
}

func (mongo *Mongo) UpdateMany(collection string, filter interface{}, update interface{}) (result *_mongo.UpdateResult, err error) {

	result, err = mongo.Database.Collection(collection).UpdateMany(
		context.Background(),
		filter,
		update,
	)

	return
}

func (mongo *Mongo) DeleteOne(collection string, filter interface{}) (result *_mongo.DeleteResult, err error) {

	result, err = mongo.Database.Collection(collection).DeleteOne(
		context.Background(),
		filter,
	)

	return
}

func (mongo *Mongo) DeleteMany(collection string, filter interface{}) (result *_mongo.DeleteResult, err error) {

	result, err = mongo.Database.Collection(collection).DeleteMany(
		context.Background(),
		filter,
	)

	return
}

func (mongo *Mongo) CountDocuments(collection string, filter interface{}) (count int64, err error) {

	count, err = mongo.Database.Collection(collection).CountDocuments(
		context.Background(),
		filter,
	)

	return
}

func (mongo *Mongo) ListCollectionNames() (result []string, err error) {

	result, err = mongo.Database.ListCollectionNames(context.Background(), bson.M{})
	return
}

func (mongo *Mongo) CreateCollection(coll string) (err error) {

	opts := options.CreateCollection()

	err = mongo.Database.CreateCollection(context.Background(), coll, opts)

	return
}

func (mongo *Mongo) DropCollection(coll string) (err error) {

	return mongo.Database.Collection(coll).Drop(context.Background())
}
