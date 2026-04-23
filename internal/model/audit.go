package model

import (
	"fmt"
	"time"
)

// AuditEvent describes metrics' update request after successful writing data to the storage.
type AuditEvent struct {
	Timestamp   int64    `json:"ts"`
	MetricNames []string `json:"metrics"`
	IPAddress   string   `json:"ip_address"`
}

// NewAuditEvent creates a new [AuditEvent] with [time.Now] timestamp.
func NewAuditEvent(metrics MetricSet, ipAddress string) AuditEvent {
	event := AuditEvent{
		Timestamp:   time.Now().Unix(),
		MetricNames: make([]string, 0, len(metrics)),
		IPAddress:   ipAddress,
	}

	for key := range metrics {
		event.MetricNames = append(event.MetricNames, NewAuditMetricName(key))
	}

	return event
}

// NewAuditMetricName converts given [MetricKey] to a string for audit purposes.
func NewAuditMetricName(key MetricKey) string {
	return fmt.Sprintf("%s:%s", key.Type, key.ID)
}
