package structs

import "time"

type Capabilities struct {
	CPUUsage     float64 // CPU usage percentage
	TotalMem     uint64  // Total virtual memory in bytes
	AvailableMem uint64  // Available memory in bytes
}

type Worker struct {
	ID            string    // UUID for the worker
	LastHeartbeat time.Time // Timestamp of the last heartbeat received
	Capabilities  Capabilities
	CurrentLoad   int // Current amount of virtual users assigned to the worker
	MaxLoad       int
	Available     bool // Whether the worker is available to take on more load
}
