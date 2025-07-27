package workers

import (
	"context"
	"sync"
	"time"

	"github.com/vnFuhung2903/vcs-sms/dto"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
	"go.uber.org/zap"
)

type IReportkWorker interface {
	Start(numWorkers int)
	Stop()
}

type ReportkWorker struct {
	containerService   services.IContainerService
	healthcheckService services.IHealthcheckService
	reportService      services.IReportService
	email              string
	logger             logger.ILogger
	interval           time.Duration
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 *sync.WaitGroup
}

func NewReportkWorker(
	containerService services.IContainerService,
	healthcheckService services.IHealthcheckService,
	reportService services.IReportService,
	email string,
	logger logger.ILogger,
	interval time.Duration,
) IReportkWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &ReportkWorker{
		containerService:   containerService,
		healthcheckService: healthcheckService,
		reportService:      reportService,
		email:              email,
		logger:             logger,
		interval:           interval,
		ctx:                ctx,
		cancel:             cancel,
		wg:                 &sync.WaitGroup{},
	}
}

func (w *ReportkWorker) Start(numWorkers int) {
	w.wg.Add(numWorkers)
	go w.run()
}

func (w *ReportkWorker) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *ReportkWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("daily report workers stopped")
			return
		case <-ticker.C:
			w.report()
		}
	}
}

func (w *ReportkWorker) report() {
	endTime := time.Now()
	startTime := endTime.Add(-w.interval)
	containers, total, err := w.containerService.View(w.ctx, dto.ContainerFilter{}, 1, -1, dto.ContainerSort{
		Field: "container_id", Order: dto.Asc,
	})
	if err != nil {
		w.logger.Error("failed to retrieve containers", zap.Error(err))
		return
	}

	var ids []string
	for _, container := range containers {
		ids = append(ids, container.ContainerId)
	}

	statusList, err := w.healthcheckService.GetEsStatus(w.ctx, ids, 10000, startTime, endTime, dto.Asc)
	if err != nil {
		w.logger.Error("failed to retrieve elasticsearch status", zap.Error(err))
		return
	}

	overlapStatusList, err := w.healthcheckService.GetEsStatus(w.ctx, ids, 1, endTime, time.Now(), dto.Asc)
	if err != nil {
		w.logger.Error("failed to retrieve elasticsearch status", zap.Error(err))
		return
	}

	onCount, offCount, totalUptime := w.reportService.CalculateReportStatistic(statusList, overlapStatusList, startTime, endTime)

	if err := w.reportService.SendEmail(w.ctx, w.email, int(total), onCount, offCount, totalUptime, startTime, endTime); err != nil {
		w.logger.Error("failed to email daily report", zap.Error(err))
		return
	}
	w.logger.Info("daily report emailed successfully")
}
