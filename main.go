package main

import (
	"entriesMicroService/api"
	"entriesMicroService/events"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	events.SetUp("")
	events.SetupProducer()
	events.SetUpConnAndStore()

	// conf, err := util.LoadConfig(".")
	// if err != nil {
	// 	log.Fatal("cannot load config:", err)
	// }

	// conn, err := sql.Open(conf.DBDriver, conf.DBSource)
	// if err != nil {
	// 	log.Fatal("cannot connect to db:", err)
	// }

	// store := db.NewStore(conn)

	server, err := api.NewServer(events.Conf, events.Store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(events.Conf.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
