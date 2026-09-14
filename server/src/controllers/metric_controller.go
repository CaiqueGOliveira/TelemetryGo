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

	accepted, published, err := mc.metricUsecase.Ingest(metrics, ctx.GetString("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"accepted":  accepted,
		"published": published,
	})
}

func (mc *MetricController) List(ctx *gin.Context) {
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

	filter := r.MetricFilter{
		Status:  ctx.Query("status"),
		Name:    ctx.Query("name"),
		Service: ctx.Query("service"),
		Start:   start,
		End:     end,
	}

	metrics, err := mc.metricUsecase.List(
		ctx.GetString("user_id"),
		filter,
		parseLimit(ctx.Query("limit"), defaultListLimit, maxListLimit),
		parseOffset(ctx.Query("offset")),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list metrics"})
		return
	}

	ctx.JSON(http.StatusOK, dtos.ToMetricResponseDtos(metrics))
}

func (mc *MetricController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid metric id"})
		return
	}

	if err := mc.metricUsecase.Delete(ctx.GetString("user_id"), id); err != nil {
		if errors.Is(err, r.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete metric"})
		return
	}

	ctx.Status(http.StatusNoContent)
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
