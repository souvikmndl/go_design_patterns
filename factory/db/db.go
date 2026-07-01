package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Factory struct {
	name        string
	mognoClient *mongo.Client
	sqlClient   *gorm.DB
	// other clients ...
}

type FactoryMethod interface {
	SetName(name string)
	GetName() string
	GetMongoClient() *mongo.Client
	GetSqlClient() *gorm.DB
}

func (g *Factory) SetName(name string) {
	g.name = name
}

func (g *Factory) GetName() string {
	return g.name
}

func (g *Factory) GetMongoClient() *mongo.Client {
	return g.mognoClient
}
func (g *Factory) GetSqlClient() *gorm.DB {
	return g.sqlClient
}

type sql struct {
	Factory
}

func sqlConnection() FactoryMethod {
	dsn := `host=localhost 
			user=test1 
			password=password 
			dbname=test 
			port=5432 
			sslmode=disable 
			TimeZone=Asia/Shanghai`
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("error occured")
		log.Println(err)
	}
	return &sql{
		Factory: Factory{
			name:      "sql",
			sqlClient: db,
		},
	}
}

type mongodb struct {
	Factory
}

func mongodbConnection() FactoryMethod {
	uri := "localhost:27017"
	client, err := mongo.NewClient(options.Client().ApplyURI(fmt.Sprintf("%s%s", "mongodb://", uri)))
	if err != nil {
		fmt.Println(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}
	singleInstance := &mongodb{Factory: Factory{name: "mongo", mognoClient: client}}
	return singleInstance

}

func GetDb(dbType string) (FactoryMethod, error) {
	if dbType == "sql" {
		return sqlConnection(), nil
	}
	if dbType == "mongo" {
		return mongodbConnection(), nil
	}
	return nil, fmt.Errorf("Wrong db type passed")
}
