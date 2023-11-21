package aktype

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Consumer interface {
	Consume(msg types.Message) error
}

type SQSReceiveMessage interface {
	StartConcurrent(ctx context.Context, handle Consumer)
}

type SQSReceiveMessageAPI interface {
	GetQueueUrl(ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)

	ReceiveMessage(ctx context.Context,
		params *sqs.ReceiveMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)

	DeleteMessage(ctx context.Context,
		params *sqs.DeleteMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)

	DeleteMessageBatch(ctx context.Context,
		params *sqs.DeleteMessageBatchInput,
		optFns ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error)
}
