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

	for mkey := range metrics.Keys() {
		event.MetricNames = append(event.MetricNames, fmt.Sprintf("%s:%s", mkey.Type, mkey.ID))
	}

	return event
}
