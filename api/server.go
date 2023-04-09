package api

import (
	"fmt"

	db "entriesMicroService/db/sqlc"
	"entriesMicroService/events"
	"entriesMicroService/token"
	"entriesMicroService/util"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
	consumer   *kafka.Consumer
}

// Creates a new HTTP server and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	events.SetUp("specific")
	consumer, err := kafka.NewConsumer(&events.KafkaConfig)
	if err != nil {
		fmt.Printf("Failed to create consumer: %s", err)
	}
	err = consumer.SubscribeTopics([]string{"user_topic"}, nil)
	if err != nil {
		fmt.Printf("Failed to subscribe to topic: %s", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	server.consumer = consumer

	server.setUpRouter()
	return server, nil
}

func (server *Server) setUpRouter() {
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	authRoutes := router.Group("/api/entry").Use(authMiddleware(server.tokenMaker))
	authRoutes.POST("/", server.addEntry)
	authRoutes.PATCH("/updateEntry", server.updateEntry)
	authRoutes.DELETE("/deleteEntry/:id", server.deleteEntry)
	authRoutes.GET("/entries", server.getEntries)
	authRoutes.GET("/categories", server.getCategories)

	server.router = router
}

// Runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// Delete this function
// func errorResponse(err error) gin.H {
// 	return gin.H{"error": err.Error()}
// }
