package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/aaslamin/link-checker/linkworker"
)

var (
	WorkerNotFoundErr  = errors.New("no worker exists with that id")
	DuplicateWorkerErr = errors.New("worker already exists with that id")
)

type MemoryStorage struct {
	Workers map[string][]linkworker.LinkError
	mux     sync.Mutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Workers: make(map[string][]linkworker.LinkError),
	}
}

func (s *MemoryStorage) AddWorker(ctx context.Context, workerID string) error {
	_, exists := s.Workers[workerID]
	if exists {
		return DuplicateWorkerErr
	}

	s.Workers[workerID] = []linkworker.LinkError{}
	return nil
}

func (s *MemoryStorage) RemoveWorker(ctx context.Context, workerID string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, exists := s.Workers[workerID]
	if !exists {
		return WorkerNotFoundErr
	}

	delete(s.Workers, workerID)
	return nil
}

func (s *MemoryStorage) GetWorkerResults(ctx context.Context, workerID string) ([]linkworker.LinkError, error) {
	results, exists := s.Workers[workerID]
	if !exists {
		return nil, WorkerNotFoundErr
	}

	return results, nil
}

func (s *MemoryStorage) AddLinkError(ctx context.Context, workerID string, linkError linkworker.LinkError) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, exists := s.Workers[workerID]
	if !exists {
		return WorkerNotFoundErr
	}

	s.Workers[workerID] = append(s.Workers[workerID], linkError)
	return nil
}
