package queue

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/mailgo/backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Adapter struct {
	client *asynq.Client
}

func NewAdapter(redisClient *redis.Client) *Adapter {
	return &Adapter{
		client: asynq.NewClient(asynq.RedisClientOpt{
			Addr: redisClient.Options().Addr,
			DB:   redisClient.Options().DB,
		}),
	}
}

func (a *Adapter) EnqueueMailboxProvision(ctx context.Context, mailboxID uuid.UUID) error {
	payload, _ := json.Marshal(map[string]interface{}{"mailbox_id": mailboxID})
	task := asynq.NewTask("mailbox:provision", payload)
	_, err := a.client.Enqueue(task, asynq.Queue("default"))
	return err
}

func (a *Adapter) EnqueueMailboxDelete(ctx context.Context, email string) error {
	payload, _ := json.Marshal(map[string]interface{}{"email": email})
	task := asynq.NewTask("mailbox:delete", payload)
	_, err := a.client.Enqueue(task, asynq.Queue("default"))
	return err
}

func (a *Adapter) EnqueueDomainVerification(ctx context.Context, domainID uuid.UUID) error {
	payload, _ := json.Marshal(map[string]interface{}{"domain_id": domainID})
	task := asynq.NewTask("domain:verify", payload)
	_, err := a.client.Enqueue(task, asynq.Queue("low"))
	return err
}

func (a *Adapter) EnqueueEmailSend(ctx context.Context, req domain.SendEmailRequest) error {
	payload, _ := json.Marshal(req)
	task := asynq.NewTask("email:send", payload)
	_, err := a.client.Enqueue(task, asynq.Queue("critical"))
	return err
}

func (a *Adapter) EnqueueAuditLogProcess(ctx context.Context, logID uuid.UUID) error {
	payload, _ := json.Marshal(map[string]interface{}{"log_id": logID})
	task := asynq.NewTask("audit:process", payload)
	_, err := a.client.Enqueue(task, asynq.Queue("low"))
	return err
}
