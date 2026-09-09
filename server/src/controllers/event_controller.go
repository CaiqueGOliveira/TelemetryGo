package controllers

import (
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
