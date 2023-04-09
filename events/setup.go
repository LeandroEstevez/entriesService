package events

import (
	"database/sql"
	db "entriesMicroService/db/sqlc"
	"entriesMicroService/util"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var KafkaConfig kafka.ConfigMap

func SetUp(groupId string) {
	KafkaConfig = util.LoadKafkaConfig("kafkaConfig.properties", groupId)
}

var Store db.Store
var Conf util.Config

func SetUpConnAndStore() {
	var err error
	Conf, err = util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(Conf.DBDriver, Conf.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	Store = db.NewStore(conn)
}
