package managers

import (
    "sync"
    "time"
    "github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/pkg/structs"
)

var (
    workers map[string]*structs.Worker
    lock    sync.RWMutex
)

func init() {
    workers = make(map[string]*structs.Worker)
}

func GetWorker(workerID string) (*structs.Worker, bool) {
	lock.RLock()
	defer lock.RUnlock()

	worker, exists := workers[workerID]
	return worker, exists
}

func GetAvailableWorkers() []*structs.Worker {
    lock.RLock()
    defer lock.RUnlock()

    var availableWorkers []*structs.Worker
    now := time.Now()
    for _, worker := range workers {
        if now.Sub(worker.LastHeartbeat) <= time.Second*40 && worker.Available {
            availableWorkers = append(availableWorkers, worker)
        }
    }
    return availableWorkers
}

func AddOrUpdateWorker(workerID string, worker *structs.Worker) {
    lock.Lock()
    defer lock.Unlock()

    workers[workerID] = worker
}

func RemoveWorker(workerID string) {
    lock.Lock()
    defer lock.Unlock()

    delete(workers, workerID)
}