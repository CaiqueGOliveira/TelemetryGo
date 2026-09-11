package controllers

import (
	"fmt"
	"net/http"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/gin-gonic/gin"
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

	accepted, err := ec.eventUsecase.Ingest(events, ctx.GetString("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"accepted": accepted})
}

func (ec *EventController) List(ctx *gin.Context) {
	events := ec.eventUsecase.List(ctx.GetString("user_id"))

	ctx.JSON(http.StatusOK, dtos.ToEventResponseDtos(events))
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
