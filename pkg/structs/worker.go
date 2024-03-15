package structs

import "time"

type Capabilities struct {
	CPUUsage     float64 `json:"cpuUsage"`
	TotalMem     uint64  `json:"totalMem"` // in bytes
	AvailableMem uint64  `json:"availableMem"`
}

type Worker struct {
	ID            string       `json:"id"`          // UUID for the worker
	LastHeartbeat time.Time    `json:lastHeartbeat` // Timestamp of the last heartbeat received
	Capabilities  Capabilities `json:"capabilities"`
	CurrentLoad   int          `json:"currentLoad"` // Current amount of virtual users assigned to the worker
	MaxLoad       int          `json:"maxLoad"`
	Available     bool         `json:"available"` // Whether the worker is available to take on more load
}
