package workers

import (
	"context"
	"sync"
	"time"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/pkg/docker"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
	"go.uber.org/zap"
)

type IHealthcheckWorker interface {
	Start(numWorkers int)
	Stop()
}

type HealthcheckWorker struct {
	dockerClient       docker.IDockerClient
	containerService   services.IContainerService
	healthcheckService services.IHealthcheckService
	logger             logger.ILogger
	interval           time.Duration
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 *sync.WaitGroup
}

func NewHealthcheckWorker(
	dockerClient docker.IDockerClient,
	containerService services.IContainerService,
	healthcheckService services.IHealthcheckService,
	logger logger.ILogger,
	interval time.Duration,
) IHealthcheckWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthcheckWorker{
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

func (w *HealthcheckWorker) Start(numWorkers int) {
	w.wg.Add(numWorkers)
	go w.run()
}

func (w *HealthcheckWorker) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *HealthcheckWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("elasticsearch status workers stopped")
			return
		case <-ticker.C:
			w.updateHealthcheck()
		}
	}
}

func (w *HealthcheckWorker) updateHealthcheck() {
	containers, total, err := w.containerService.View(w.ctx, dto.ContainerFilter{}, 1, -1, dto.ContainerSort{
		Field: "updated_at", Order: "desc",
	})
	if err != nil {
		w.logger.Error("failed to view containers", zap.Error(err))
		return
	}

	statusList := make([]dto.EsStatusUpdate, 0, total)

	for _, container := range containers {
		status := w.dockerClient.GetStatus(w.ctx, container.ContainerId)
		if status != container.Status {
			if err := w.containerService.Update(w.ctx, container.ContainerId, dto.ContainerUpdate{Status: status}); err != nil {
				w.logger.Error("failed to update container", zap.String("container_id", container.ContainerId))
			}
		}
		statusList = append(statusList, dto.EsStatusUpdate{
			ContainerId: container.ContainerId,
			Status:      status,
		})
	}

	if err := w.healthcheckService.UpdateStatus(w.ctx, statusList); err != nil {
		w.logger.Error("failed to update elasticsearch status", zap.Error(err))
		return
	}
	w.logger.Info("elasticsearch status updated successfully")
}
