package health

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type Checker struct {
	postgres *pgxpool.Pool
	redis    *redis.Client
	rabbitMQ *amqp091.Connection
}

func NewChecker(
	postgres *pgxpool.Pool,
	redisClient *redis.Client,
	rabbitMQ *amqp091.Connection,
) *Checker {
	return &Checker{
		postgres: postgres,
		redis:    redisClient,
		rabbitMQ: rabbitMQ,
	}
}

func (c *Checker) CheckPostgres(ctx context.Context) error {
	return c.postgres.Ping(ctx)
}

func (c *Checker) CheckRedis(ctx context.Context) error {
	return c.redis.Ping(ctx).Err()
}

func (c *Checker) CheckRabbitMQ(ctx context.Context) error {
	if c.rabbitMQ == nil || c.rabbitMQ.IsClosed() {
		return amqp091.ErrClosed
	}

	return nil
}
