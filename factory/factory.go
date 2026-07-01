package factory

import (
	"fmt"
	database "go_design_patterns/factory/db"
)

func FacotryStart() {
	sql, _ := database.GetDb("sql")
	mongo, _ := database.GetDb("mongo")

	printDetails(sql)
	printDetails(mongo)
}

func printDetails(g database.FactoryMethod) {
	fmt.Printf("\n")
	fmt.Printf("db: %s", g.GetName())
	if g.GetName() == "sql" {
		fmt.Printf("\n")
		fmt.Printf("client: %v", g.GetSqlClient())
		fmt.Printf("\n")
	} else {
		fmt.Printf("\n")
		fmt.Printf("client: %v\n", g.GetMongoClient())
		fmt.Printf("\n")
	}
}
