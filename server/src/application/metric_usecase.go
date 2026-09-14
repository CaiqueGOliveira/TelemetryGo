package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/messages"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

const MetricsChannel = "telemetry:metrics"

type MetricUsecase struct {
	repo      r.MetricRepository
	publisher messages.EventPublisher
}

func NewMetricUsecase(repo r.MetricRepository, publisher messages.EventPublisher) *MetricUsecase {
	return &MetricUsecase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *MetricUsecase) Ingest(requests []*dtos.MetricIngestRequestDto, userID string) (int, int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var accepted int
	var published int
	var firstErr error

	for _, req := range requests {
		wg.Add(1)
		go func(req *dtos.MetricIngestRequestDto) {
			defer wg.Done()

			metric, err := buildMetric(req, userID)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}

			if err := uc.repo.Save(metric); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}

			accepted++

			payload, err := json.Marshal(dtos.ToMetricResponseDto(metric))
			if err != nil {
				slog.Error("failed to marshal metric for publish", "metric_id", metric.Id, "error", err)
				return
			}

			if err := uc.publisher.Publish(context.Background(), MetricsChannel, payload); err != nil {
				slog.Error("failed to publish metric", "channel", MetricsChannel, "error", err)
				return
			}

			published++
		}(req)
	}

	wg.Wait()

	if firstErr != nil {
		return 0, 0, firstErr
	}

	return accepted, published, nil
}

func (uc *MetricUsecase) List(userID string, filter r.MetricFilter, limit int, offset int) ([]*domain.Metric, error) {
	return uc.repo.List(userID, filter, limit, offset)
}

func (uc *MetricUsecase) Delete(userID string, id uuid.UUID) error {
	return uc.repo.Delete(userID, id)
}

func (uc *MetricUsecase) Subscribe(ctx context.Context) (<-chan []byte, func(), error) {
	return uc.publisher.Subscribe(ctx, MetricsChannel)
}

func buildMetric(req *dtos.MetricIngestRequestDto, userID string) (*domain.Metric, error) {
	id := uuid.New()
	if req.ID != "" {
		parsedID, err := uuid.Parse(req.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid metric id: %w", err)
		}
		id = parsedID
	}

	if err := validateRequiredMetricFields(req); err != nil {
		return nil, err
	}

	value, err := strconv.ParseFloat(req.Value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid metric value %q: numeric value required", req.Value)
	}

	timestamp := time.Now()
	if req.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, req.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("invalid metric timestamp: %w", err)
		}
		timestamp = parsed
	}

	metric, err := domain.NewMetric(
		id,
		req.Name,
		req.Service,
		value,
		req.Unit,
		req.Status,
		timestamp,
	)
	if err != nil {
		return nil, err
	}

	metric.UserId = userID
	return metric, nil
}

func validateRequiredMetricFields(req *dtos.MetricIngestRequestDto) error {
	switch {
	case req.Name == "":
		return fmt.Errorf("metric name is required")
	case req.Service == "":
		return fmt.Errorf("metric service is required")
	case req.Value == "":
		return fmt.Errorf("metric value is required")
	default:
		return nil
	}
}
