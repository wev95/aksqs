package aksqs

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Queue interface {
	Send(ctx context.Context, queue string, message interface{}) error
}

type SqsQueue struct {
	client *sqs.Client
}

func NewQueue(client *sqs.Client) Queue {
	return &SqsQueue{client: client}
}

func (q *SqsQueue) Send(ctx context.Context, queue string, message interface{}) error {

	urlResult, err := q.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String(queue)})

	if err != nil {
		return err
	}

	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = q.client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(string(msg)),
		QueueUrl:    urlResult.QueueUrl,
	})

	return err
}
