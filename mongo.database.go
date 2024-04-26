package Grizzly

import (
	"context"
	"log"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDBs map[string]*MongoDB

type MongoDBStatus struct {
	Connected bool
}

type MongoDBConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	AuthDatabase string
}

type MongoDB struct {
	Config   MongoDBConfig
	client   *mongo.Client
	Database *mongo.Database
}

func (db *MongoDB) createConnectionString(MongoDBConfig) string {
	res := "mongodb://"
	if db.Config.Username != "" {
		res += db.Config.Username + ":" + db.Config.Password + "@"
	}
	res += db.Config.Host + ":" + strconv.Itoa(db.Config.Port)
	res += "/" + db.Config.Database
	if db.Config.AuthDatabase != "" {
		res += "?authSource=" + db.Config.AuthDatabase
	}
	return res
}

func (db *MongoDB) connect() error {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(db.createConnectionString(db.Config)))
	if err != nil {
		return err
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}
	db.client = client
	db.Database = db.client.Database(db.Config.Database)
	return nil
}

func (db *MongoDB) GetStatus() MongoDBStatus {
	return MongoDBStatus{
		Connected: db.client != nil,
	}
}

func (db *MongoDB) GetClient() *mongo.Client {
	return db.client
}

func (db *MongoDB) Reconnect() *mongo.Database {
	err := db.client.Disconnect(context.Background())
	if err != nil {
		panic(err)
	}
	err = db.connect()
	if err != nil {
		panic(err)
	}
	return db.client.Database(db.Config.Database)

}

func InitDBConnection(connectionName string, config MongoDBConfig) {
	if mongoDBs == nil {
		mongoDBs = make(map[string]*MongoDB)
	}
	db := MongoDB{
		Config: config,
	}
	err := db.connect()
	if err != nil {
		panic(err)
	} else {
		// log.Print("Connected to MongoDB at", db.Config.Host+":"+strconv.Itoa(db.Config.Port), "using database \"", db.Config.Database)
		log.Printf("Connected to MongoDB at \"%s:%d\" using database \"%s\"", db.Config.Host, db.Config.Port, db.Config.Database)
	}
	mongoDBs[connectionName] = &db
}

func GetDBConnection(connectionName string) *MongoDB {
	return mongoDBs[connectionName]
}
