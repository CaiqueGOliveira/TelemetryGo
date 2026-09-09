package application

import (
	"sync"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/application/dtos"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	r "github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

type MetricUsecase struct {
	repo r.MetricRepository
}

func NewMetricUsecase(repo r.MetricRepository) *MetricUsecase {
	return &MetricUsecase{
		repo: repo,
	}
}

func (uc *MetricUsecase) Ingest(requests []*dtos.MetricIngestRequestDto, userID string) (int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var accepted int
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
		}(req)
	}

	wg.Wait()

	if firstErr != nil {
		return 0, firstErr
	}

	return accepted, nil
}

func (uc *MetricUsecase) List(userID string) []*domain.Metric {
	return uc.repo.FindAll(userID)
}

func buildMetric(req *dtos.MetricIngestRequestDto, userID string) (*domain.Metric, error) {
	id := req.ID
	if id == "" {
		id = uuid.NewString()
	}

	name := req.Name
	if name == "" {
		name = "metric"
	}

	timestamp := time.Now()
	if req.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, req.Timestamp)
		if err == nil {
			timestamp = parsed
		}
	}

	metric, err := domain.NewMetric(
		id,
		name,
		req.Service,
		req.Value,
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
