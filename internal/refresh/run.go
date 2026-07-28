package refresh

import (
	"context"
	"errors"
	"sync"

	repositoryScope "github.com/mhmdnurf/github-stats/internal/repository"
)

const maxConcurrentFetches = 2

type refreshTask func(context.Context) error

func (service *Service) Run(ctx context.Context) error {
	tasks := []refreshTask{
		func(ctx context.Context) error {
			return service.refreshStats(
				ctx,
				repositoryScope.ScopePublic,
			)
		},
		func(ctx context.Context) error {
			return service.refreshStats(
				ctx,
				repositoryScope.ScopeAll,
			)
		},
		func(ctx context.Context) error {
			return service.refreshLanguages(
				ctx,
				repositoryScope.ScopePublic,
			)
		},
		func(ctx context.Context) error {
			return service.refreshLanguages(
				ctx,
				repositoryScope.ScopeAll,
			)
		},
	}

	semaphore := make(
		chan struct{},
		maxConcurrentFetches,
	)
	errorsChannel := make(chan error, len(tasks))
	var waitGroup sync.WaitGroup

	for _, runTask := range tasks {
		runTask := runTask
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			select {
			case semaphore <- struct{}{}:
				defer func() {
					<-semaphore
				}()
			case <-ctx.Done():
				errorsChannel <- ctx.Err()
				return
			}

			if err := runTask(ctx); err != nil {
				errorsChannel <- err
			}
		}()
	}

	waitGroup.Wait()
	close(errorsChannel)

	var refreshErrors []error
	for err := range errorsChannel {
		refreshErrors = append(refreshErrors, err)
	}

	return errors.Join(refreshErrors...)
}
