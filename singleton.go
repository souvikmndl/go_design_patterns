package main

// Welcome to your channel go guru

// Topic design patterns

// What is design patterns:-
// reusable solution for the common problems in software design
// typically shows relationships and interactions between classes or objects
// speed up the development
// provide us a well tested code
// independent to programing language. That means a design pattern represents an idea, not a particular implementation.
// design patterns make your code more flexible, reusable, and maintainable.

// Types of Design Patterns
// Creational -> signleton method, Factory Method, Abstract Factory, etc.
// Structural -> Adapter, Bridge, etc.
// Behavioral -> Command, Interpreter, etc.

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var lock = &sync.Mutex{}

// 1000 => create user {}

type single struct {
	connection *mongo.Client
}

var singleInstance *single

func getInstance() *single {
	lock.Lock()
	defer lock.Unlock()
	if singleInstance == nil {
		fmt.Println("Creating db connection")
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
		fmt.Println("Connected!!!")
		singleInstance = &single{connection: client}
	} else {
		fmt.Println("Db connection is already created.")
	}
	return singleInstance
}

func singletonExample() {

	for i := 0; i < 30; i++ {
		go getInstance()
	}

	// Scanln is similar to Scan, but stops scanning at a newline and
	// after the final item there must be a newline or EOF.
	fmt.Scanln()
}
