package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/lib/pq"
)

func Listen(message *kafka.Message) {
	key := string(message.Key)

	switch key {
	case "user_deleted":
		UserDeleted(message.Value)
	}
}

type DefaultMessage struct {
	Value string `json:"value"`
}

func UserDeleted(value []byte) {
	var msg DefaultMessage
	json.Unmarshal(value, &msg)

	fmt.Println(msg)
	err := Store.DeleteEntries(context.TODO(), msg.Value)
	if err != nil {
		fmt.Println("Entries were not deleted")
		return
		// TODO: send back an event to entriesMicroService with the error
	}
	fmt.Println("Entries were deleted")
}
