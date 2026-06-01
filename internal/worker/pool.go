package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/VarvaraKurakova/fleet-events-api/internal/messaging"
	"github.com/VarvaraKurakova/fleet-events-api/internal/service"
)

type Pool struct {
	logger       *slog.Logger
	alertService *service.AlertService
	concurrency  int
	jobs         chan amqp.Delivery
	wg           sync.WaitGroup
}

func NewPool(
	logger *slog.Logger,
	alertService *service.AlertService,
	concurrency int,
) *Pool {
	if concurrency <= 0 {
		concurrency = 1
	}

	return &Pool{
		logger:       logger,
		alertService: alertService,
		concurrency:  concurrency,
		jobs:         make(chan amqp.Delivery),
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.concurrency; i++ {
		workerID := i + 1

		p.wg.Add(1)
		go p.runWorker(ctx, workerID)
	}
}

func (p *Pool) Submit(ctx context.Context, delivery amqp.Delivery) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.jobs <- delivery:
		return nil
	}
}

func (p *Pool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}

func (p *Pool) runWorker(ctx context.Context, workerID int) {
	defer p.wg.Done()

	p.logger.Info("worker goroutine started", "worker_id", workerID)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("worker goroutine stopping", "worker_id", workerID)
			return

		case delivery, ok := <-p.jobs:
			if !ok {
				p.logger.Info("worker jobs channel closed", "worker_id", workerID)
				return
			}

			p.handleDelivery(ctx, workerID, delivery)
		}
	}
}

func (p *Pool) handleDelivery(ctx context.Context, workerID int, delivery amqp.Delivery) {
	if err := p.processDelivery(ctx, workerID, delivery); err != nil {
		p.logger.Error(
			"failed to process message",
			"worker_id", workerID,
			"error", err,
		)

		if isPermanentMessageError(err) {
			if rejectErr := delivery.Reject(false); rejectErr != nil {
				p.logger.Error(
					"failed to reject message",
					"worker_id", workerID,
					"error", rejectErr,
				)
			}

			return
		}

		if nackErr := delivery.Nack(false, true); nackErr != nil {
			p.logger.Error(
				"failed to nack message",
				"worker_id", workerID,
				"error", nackErr,
			)
		}

		return
	}

	if err := delivery.Ack(false); err != nil {
		p.logger.Error(
			"failed to ack message",
			"worker_id", workerID,
			"error", err,
		)
	}
}

func (p *Pool) processDelivery(ctx context.Context, workerID int, delivery amqp.Delivery) error {
	processCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var message messaging.EventCreatedMessage
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
	}

	alerts, err := p.alertService.ProcessEvent(processCtx, message)
	if err != nil {
		return err
	}

	p.logger.Info(
		"event processed",
		"worker_id", workerID,
		"event_id", message.EventID,
		"vehicle_id", message.VehicleID,
		"alerts_created", len(alerts),
	)

	return nil
}

var ErrInvalidMessage = errors.New("invalid message")

func isPermanentMessageError(err error) bool {
	return errors.Is(err, ErrInvalidMessage)
}
