package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventController struct {
	eventUsecase *application.EventUsecase
}

func NewEventController(eventUsecase *application.EventUsecase) *EventController {
	return &EventController{
		eventUsecase: eventUsecase,
	}
}

func (ec *EventController) Ingest(ctx *gin.Context) {
	var events []*dtos.EventIngestRequestDto

	if err := ctx.ShouldBindJSON(&events); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if len(events) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no events provided"})
		return
	}

	accepted, published, err := ec.eventUsecase.Ingest(events, ctx.GetString("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"accepted":  accepted,
		"published": published,
	})
}

const (
	defaultListLimit = 50
	maxListLimit     = 100
)

func (ec *EventController) List(ctx *gin.Context) {
	start, err := parseTime(ctx.Query("start"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	end, err := parseTime(ctx.Query("end"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := r.EventFilter{
		Severity: ctx.Query("severity"),
		Type:     ctx.Query("type"),
		Service:  ctx.Query("service"),
		Start:    start,
		End:      end,
	}

	events, err := ec.eventUsecase.List(
		ctx.GetString("user_id"),
		filter,
		parseLimit(ctx.Query("limit"), defaultListLimit, maxListLimit),
		parseOffset(ctx.Query("offset")),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}

	ctx.JSON(http.StatusOK, dtos.ToEventResponseDtos(events))
}

func (ec *EventController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	if err := ec.eventUsecase.Delete(ctx.GetString("user_id"), id); err != nil {
		if errors.Is(err, r.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (ec *EventController) Stream(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	ch, cleanup, err := ec.eventUsecase.Subscribe(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe to events"})
		return
	}
	defer cleanup()

	ctx.Writer.Flush()

	ctx.SSEvent("connected", "listening for events")
	ctx.Writer.Flush()

	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		case payload, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(ctx.Writer, "data: %s\n\n", payload)
			ctx.Writer.Flush()
		}
	}
}
