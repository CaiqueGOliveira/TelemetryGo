package valueobjects

import "errors"

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

func CreateSeverity(value string) (Severity, error) {
	severity := Severity(value)
	switch severity {
	case SeverityInfo, SeverityWarning, SeverityCritical:
		return severity, nil
	default:
		return "", errors.New("invalid severity level")
	}
}

func (s Severity) String() string {
	return string(s)
}
