package sqsqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/voxire/onion/internal/platform/workerqueue"
)

type Queue struct {
	client            *sqs.Client
	name              string
	queueURL          string
	waitTimeSeconds   int32
	visibilitySeconds int32
}

type Option func(*Queue)

func WithWaitTime(seconds int32) Option {
	return func(q *Queue) {
		if seconds >= 0 {
			q.waitTimeSeconds = seconds
		}
	}
}

func WithVisibilityTimeout(seconds int32) Option {
	return func(q *Queue) {
		if seconds > 0 {
			q.visibilitySeconds = seconds
		}
	}
}

func New(client *sqs.Client, name, queueURL string, opts ...Option) *Queue {
	q := &Queue{
		client:            client,
		name:              name,
		queueURL:          queueURL,
		waitTimeSeconds:   10,
		visibilitySeconds: 30,
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

func (q *Queue) Name() string {
	if q.name != "" {
		return q.name
	}
	return q.queueURL
}

func (q *Queue) Receive(ctx context.Context, max int) ([]workerqueue.Message, error) {
	if q == nil || q.client == nil {
		return nil, fmt.Errorf("sqs client is required")
	}
	if q.queueURL == "" {
		return nil, fmt.Errorf("sqs queue url is required")
	}
	if max <= 0 || max > 10 {
		max = 10
	}

	out, err := q.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &q.queueURL,
		MaxNumberOfMessages: int32(max),
		WaitTimeSeconds:     q.waitTimeSeconds,
		VisibilityTimeout:   q.visibilitySeconds,
		AttributeNames:      []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if err != nil {
		return nil, err
	}

	messages := make([]workerqueue.Message, 0, len(out.Messages))
	for _, raw := range out.Messages {
		messageID := value(raw.MessageId)
		body := []byte(value(raw.Body))
		receipt := value(raw.ReceiptHandle)
		receivedAt := time.Now().UTC()
		messages = append(messages, workerqueue.Message{
			ID:         messageID,
			Body:       body,
			ReceivedAt: receivedAt,
			Ack: func(ctx context.Context) error {
				if receipt == "" {
					return nil
				}
				_, err := q.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      &q.queueURL,
					ReceiptHandle: &receipt,
				})
				return err
			},
			Nack: func(ctx context.Context, _ error) error {
				if receipt == "" {
					return nil
				}
				_, err := q.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
					QueueUrl:          &q.queueURL,
					ReceiptHandle:     &receipt,
					VisibilityTimeout: 0,
				})
				return err
			},
		})
	}
	return messages, nil
}

func value(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
