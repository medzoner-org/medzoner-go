package main

import (
	"log"
	"os"

	wiring "github.com/Medzoner/medzoner-go/internal/wire"
	"github.com/Medzoner/medzoner-go/pkg/database"
)

var migrateAction = database.Up

func main() {
	mg, err := wiring.InitDbMigration()
	if err != nil {
		panic(err)
	}
	mg.DbInstance.CreateDatabase(mg.DbInstance.GetDatabaseName())
	if len(os.Args) > 1 {
		migrateAction = os.Args[1]
	}
	if migrateAction != database.Down && migrateAction != database.Up {
		log.Fatal("Invalid flag")
	}

	if err = mg.Migrate(migrateAction); err != nil {
		log.Fatal(err)
	}
}
