package main

import (
	"log"

	"github.com/karlsburg87/statusConfig/internal/api"
)

func main() {

	manager := api.Setup()
	defer manager.Datastore.Close()

	//serve Configerations to statusSentry services
	log.Fatalln(manager.Server.ListenAndServe())
}
