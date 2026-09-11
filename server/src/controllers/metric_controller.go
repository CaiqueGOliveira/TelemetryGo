package controllers

import (
	"fmt"
	"net/http"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application"
	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/gin-gonic/gin"
)

type MetricController struct {
	metricUsecase *application.MetricUsecase
}

func NewMetricController(metricUsecase *application.MetricUsecase) *MetricController {
	return &MetricController{
		metricUsecase: metricUsecase,
	}
}

func (mc *MetricController) Ingest(ctx *gin.Context) {
	var metrics []*dtos.MetricIngestRequestDto

	if err := ctx.ShouldBindJSON(&metrics); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if len(metrics) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no metrics provided"})
		return
	}

	accepted, err := mc.metricUsecase.Ingest(metrics, ctx.GetString("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"accepted": accepted})
}

func (mc *MetricController) List(ctx *gin.Context) {
	metrics := mc.metricUsecase.List(ctx.GetString("user_id"))

	ctx.JSON(http.StatusOK, dtos.ToMetricResponseDtos(metrics))
}

func (mc *MetricController) Stream(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	ch, cleanup, err := mc.metricUsecase.Subscribe(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to subscribe to metrics"})
		return
	}
	defer cleanup()

	ctx.Writer.Flush()

	ctx.SSEvent("connected", "listening for metrics")
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
