package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	db "entriesMicroService/db/sqlc"
	"entriesMicroService/events"
	"entriesMicroService/token"
	"entriesMicroService/util"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
)

const (
	YYYYMMDD = "2006-01-02"
)

type DefaultMessage struct {
	Value string
}

func (server *Server) QuickConsume() []byte {
	var event *kafka.Message

	// Process messages
	run := true
	for run {
		var err error
		event, err = server.consumer.ReadMessage(100 * time.Millisecond)
		if err != nil {
			continue
		}
		fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
			*event.TopicPartition.Topic, string(event.Key), string(event.Value))

		break
	}

	return event.Value
}

type createEntryRequest struct {
	Username string `json:"username" binding:"required,min=6,max=10"`
	Name     string `json:"name" binding:"required,alphaunicode"`
	DueDate  string `json:"due_date" binding:"required" time_format:"2006-01-02"`
	Amount   int64  `json:"amount" binding:"required,gt=0"`
	Category string `json:"category" binding:"max=10"`
}

type updatedEntryMessage struct {
	OrignalEntry db.Entry
	UpdatedEntry db.Entry
}

func (server *Server) addEntry(ctx *gin.Context) {
	var req createEntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	dueDate, err := time.Parse(YYYYMMDD, req.DueDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.AddEntryTxParams{
		Username: authPayload.Username,
		Name:     req.Name,
		DueDate:  dueDate,
		Amount:   req.Amount,
		Category: req.Category,
	}

	entryResult, err := server.store.AddEntryTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	events.Produce("entry_created", entryResult.Entry)

	// Check if user was correctly updated, if not, execute compensating transaction
	value := server.QuickConsume()
	var msg DefaultMessage
	err = json.Unmarshal(value, &msg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if msg.Value == "N" {
		params := db.DeleteEntryTxParams{
			Username: arg.Username,
			ID:       entryResult.Entry.ID,
		}
		_, err := server.store.DeleteEntryTx(ctx, params)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		} else {
			ctx.JSON(http.StatusInternalServerError, "Unable to update user")
		}
	} else {
		ctx.JSON(http.StatusOK, entryResult)
	}
}

type deleteEntryRequest struct {
	ID int32 `uri:"id" binding:"required,gt=0"`
}

func (server *Server) deleteEntry(ctx *gin.Context) {
	var req deleteEntryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.DeleteEntryTxParams{
		Username: authPayload.Username,
		ID:       req.ID,
	}

	compArg := db.GetEntryParams{
		Owner: authPayload.Username,
		ID:    req.ID,
	}
	compEntry, err := server.store.GetEntry(ctx, compArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	deleteEntryResult, err := server.store.DeleteEntryTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	events.Produce("entry_deleted", deleteEntryResult)

	// Check if user was correctly updated, if not, execute compensating transaction
	value := server.QuickConsume()
	var ev DefaultMessage
	err = json.Unmarshal(value, &ev)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if ev.Value == "N" {
		arg := db.CreateEntryParams{
			Owner:    compEntry.Owner,
			Name:     compEntry.Name,
			DueDate:  compEntry.DueDate,
			Amount:   compEntry.Amount,
			Category: compEntry.Category,
		}
		_, err := server.store.CreateEntry(ctx, arg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		} else {
			ctx.JSON(http.StatusInternalServerError, "Unable to update user")
		}
	} else {
		ctx.JSON(http.StatusOK, deleteEntryResult)
	}
}

type getEntriesRequest struct {
	Username string `form:"username" binding:"required,min=6,max=10"`
}

func (server *Server) getEntries(ctx *gin.Context) {
	var req getEntriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	entries, err := server.store.GetEntries(ctx, authPayload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entries)
}

func (server *Server) getCategories(ctx *gin.Context) {
	categories, err := server.store.GetCategories(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, categories)
}

type updateEntryRequest struct {
	Username string `json:"username" binding:"required,min=6,max=10"`
	ID       int32  `json:"id" binding:"required,gt=0"`
	Name     string `json:"name" binding:"required,alphaunicode"`
	DueDate  string `json:"due_date" binding:"required" time_format:"2006-01-02"`
	Amount   int64  `json:"amount" binding:"required,gt=0"`
	Category string `json:"Category" binding:"max=10"`
}

func (server *Server) updateEntry(ctx *gin.Context) {
	var req updateEntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	dueDate, err := time.Parse(YYYYMMDD, req.DueDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.UpdateEntryTxParams{
		Username: authPayload.Username,
		ID:       req.ID,
		Name:     req.Name,
		DueDate:  dueDate,
		Amount:   req.Amount,
		Category: req.Category,
	}

	compArg := db.GetEntryParams{
		Owner: authPayload.Username,
		ID:    req.ID,
	}
	compEntry, err := server.store.GetEntry(ctx, compArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	updateEntryResult, err := server.store.UpdateEntryTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	msg := updatedEntryMessage{
		OrignalEntry: updateEntryResult.Original,
		UpdatedEntry: updateEntryResult.Entry,
	}
	events.Produce("entry_updated", msg)

	// Check if user was correctly updated, if not, execute compensating transaction
	value := server.QuickConsume()
	var ev DefaultMessage
	err = json.Unmarshal(value, &ev)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if ev.Value == "N" {
		arg := db.UpdateEntryTxParams{
			Username: compEntry.Owner,
			ID:       compEntry.ID,
			Name:     compEntry.Name,
			DueDate:  compEntry.DueDate,
			Amount:   compEntry.Amount,
			Category: compEntry.Category.String,
		}
		_, err := server.store.UpdateEntryTx(ctx, arg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		} else {
			ctx.JSON(http.StatusInternalServerError, "Unable to update user")
		}
	} else {
		ctx.JSON(http.StatusOK, updateEntryResult.Entry)
	}
}
