package linkworker

import "context"

type WorkerResultStorage interface {
	AddWorker(ctx context.Context, workerID string) error
	RemoveWorker(ctx context.Context, workerID string) error
	AddLinkError(ctx context.Context, workerID string, linkError LinkError) error
	GetWorkerResults(ctx context.Context, workerID string) ([]LinkError, error)
}
