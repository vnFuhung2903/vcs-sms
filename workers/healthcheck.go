package workers

import (
	"context"
	"sync"
	"time"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/pkg/docker"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
)

type EsStatusWorker struct {
	dockerClient       docker.IDockerClient
	containerService   services.IContainerService
	healthcheckService services.IHealthcheckService
	logger             logger.ILogger
	interval           time.Duration
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 *sync.WaitGroup
}

func (w *EsStatusWorker) Start(numWorkers int) {
	w.wg.Add(numWorkers)
	go w.run()
}

func (w *EsStatusWorker) Stop() {
	w.cancel()
	w.wg.Wait()
}

func NewEsStatusWorker(
	dockerClient docker.IDockerClient,
	containerService services.IContainerService,
	healthcheckService services.IHealthcheckService,
	logger logger.ILogger,
	interval time.Duration,
) *EsStatusWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &EsStatusWorker{
		dockerClient:       dockerClient,
		healthcheckService: healthcheckService,
		containerService:   containerService,
		logger:             logger,
		interval:           interval,
		ctx:                ctx,
		cancel:             cancel,
		wg:                 &sync.WaitGroup{},
	}
}

func (w *EsStatusWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	var statusList []dto.EsStatusUpdate

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("elasticsearch status workers stopped")
			return
		case <-ticker.C:
			w.updateEsStatus(statusList)
		}
	}
}

func (w *EsStatusWorker) updateEsStatus(statusList []dto.EsStatusUpdate) {
	containers, _, err := w.containerService.View(w.ctx, dto.ContainerFilter{}, 1, -1, dto.ContainerSort{Field: "created_at", Order: "desc"})
	if err != nil {
		return
	}

	statusList = make([]dto.EsStatusUpdate, 0, len(containers))
	for _, container := range containers {
		status := w.dockerClient.GetStatus(w.ctx, container.ContainerId)
		if status != container.Status {
			w.containerService.Update(w.ctx, container.ContainerId, dto.ContainerUpdate{Status: status})
		}
		statusList = append(statusList, dto.EsStatusUpdate{
			ContainerId: container.ContainerId,
			Status:      status,
		})
	}

	err = w.healthcheckService.UpdateStatus(w.ctx, statusList)
	if err != nil {
		return
	}
	w.logger.Info("elasticsearch status updated successfully")
}
