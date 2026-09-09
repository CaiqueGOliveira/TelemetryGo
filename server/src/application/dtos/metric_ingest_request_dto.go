package dtos

type MetricIngestRequestDto struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Service   string `json:"service"`
	Value     string `json:"value"`
	Unit      string `json:"unit"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp,omitempty"`
}

func NewMetricIngestRequestDto(id string, name string, service string, value string, unit string, status string, timestamp string) *MetricIngestRequestDto {
	return &MetricIngestRequestDto{
		ID:        id,
		Name:      name,
		Service:   service,
		Value:     value,
		Unit:      unit,
		Status:    status,
		Timestamp: timestamp,
	}
}
