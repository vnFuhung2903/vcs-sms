package dto

import (
	"time"

	"github.com/vnFuhung2903/vcs-sms/entities"
)

type ReportRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Email     string    `json:"email"`
}

type ReportResponse struct {
	ContainerCount    int     `json:"container_count"`
	ContainerOnCount  int     `json:"container_on_count"`
	ContainerOffCount int     `json:"container_off_count"`
	AverageUptime     float64 `json:"average_uptime"`
}

type EsStatus struct {
	ContainerId string                   `json:"server_id"`
	Status      entities.ContainerStatus `json:"status"`
	Uptime      int64                    `json:"uptime"`
	LastUpdated time.Time                `json:"last_updated"`
}

type EsStatusUpdate struct {
	ContainerId string                   `json:"server_id"`
	Status      entities.ContainerStatus `json:"status"`
}
