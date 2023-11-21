package aksqs

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
	aktype "github.com/wev95/aksqs/type"
)

var ErrConsumerNotImplemented = errors.New("consumer is not implemented or nil")

type SQSReceive struct {
	client   aktype.SQSReceiveMessageAPI
	consumer aktype.Consumer
	gMInput  *sqs.ReceiveMessageInput
	option   SQSReceiveOption
}

type SQSReceiveOption struct {
	queueUrl            string
	queueName           string
	concurrentConsumers int
	concurrentBuff      int
	maxNumberOfMessages int32
	visibilityTimeout   int32
	waitTimeSeconds     int32
}

type SQSReceiveConfig struct {
	ConcurrentConsumers int
	ConcurrentBuff      int
	MaxNumberOfMessages int32
	VisibilityTimeout   int32
	WaitTimeSeconds     int32
	QueueName           string
}

func NewSQSReceive(ctx context.Context, c aktype.SQSReceiveMessageAPI, opt *SQSReceiveConfig) aktype.SQSReceiveMessage {

	config := opt

	if config == nil {
		config = &SQSReceiveConfig{
			ConcurrentConsumers: 1,
			ConcurrentBuff:      0,
			MaxNumberOfMessages: 1,
			VisibilityTimeout:   60,
		}
	}

	queueName := config.QueueName

	sqsReceive := SQSReceive{
		client: c,
		option: SQSReceiveOption{
			queueName:           queueName,
			concurrentConsumers: config.ConcurrentConsumers,
			concurrentBuff:      config.ConcurrentBuff,
			maxNumberOfMessages: config.MaxNumberOfMessages,
			visibilityTimeout:   config.VisibilityTimeout,
			waitTimeSeconds:     config.WaitTimeSeconds,
		}}

	resolveQueueUrl(ctx, &sqsReceive, queueName)
	resolveReceiveMessageInput(ctx, &sqsReceive, config)

	return &sqsReceive
}

func resolveQueueUrl(ctx context.Context, sqsReceive *SQSReceive, queueName string) {

	urlResult, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String(queueName)})

	if err != nil {
		err = fmt.Errorf("%v \n queue not found: %v", err, queueName)
		log.Error().Err(err).Msg("")
		panic(err)
	}

	sqsReceive.option.queueUrl = *urlResult.QueueUrl
}

func resolveReceiveMessageInput(ctx context.Context, sqsReceive *SQSReceive, opt *SQSReceiveConfig) {

	sqsReceive.gMInput = &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            aws.String(sqsReceive.option.queueUrl),
		MaxNumberOfMessages: *aws.Int32(opt.MaxNumberOfMessages),
		VisibilityTimeout:   *aws.Int32(opt.VisibilityTimeout),
		WaitTimeSeconds:     *aws.Int32(opt.WaitTimeSeconds),
	}
}

func (r *SQSReceive) StartConcurrent(ctx context.Context, handle aktype.Consumer) {

	if handle == nil {
		logger.Error().Err(ErrConsumerNotImplemented).Msg("")
		panic(ErrConsumerNotImplemented)
	}

	r.consumer = handle
	concurrentConsumers := r.option.concurrentConsumers
	buff := concurrentConsumers * r.option.concurrentBuff
	msgChan := make(chan *sqs.ReceiveMessageOutput, buff)
	deleteChan := make(chan []types.DeleteMessageBatchRequestEntry, buff)

	var wgReceive sync.WaitGroup
	wgReceive.Add(concurrentConsumers)

	var wg sync.WaitGroup
	wg.Add(concurrentConsumers)

	for i := 0; i < concurrentConsumers; i++ {
		go receive(ctx, &wgReceive, r, r.gMInput, msgChan)
		go startConsumer(r, &wg, msgChan, deleteChan)
	}

	deleteMessage(context.TODO(), r, deleteChan)

	wgReceive.Wait()
	close(msgChan)

	wg.Wait()
	close(deleteChan)
}

func receive(
	ctx context.Context,
	wg *sync.WaitGroup,
	r *SQSReceive,
	gMInput *sqs.ReceiveMessageInput,
	msgChan chan<- *sqs.ReceiveMessageOutput,
) {

	for {
		select {
		case <-ctx.Done():
			{
				wg.Done()
				return
			}
		default:
			{
				if msgResult, err := r.client.ReceiveMessage(ctx, gMInput); err != nil {
					if err == context.Canceled {
						logger.Debug().Err(err).Msg("")
						return
					}
					logger.Err(err).Msg("")
				} else if msgResult != nil && msgResult.Messages != nil {
					msgChan <- msgResult
				}
			}
		}
	}
}

func startConsumer(r *SQSReceive,
	wg *sync.WaitGroup,
	msgIn <-chan *sqs.ReceiveMessageOutput,
	msgOut chan<- []types.DeleteMessageBatchRequestEntry,
) {

	defer wg.Done()

	for msgResult := range msgIn {

		messages := msgResult.Messages

		entries := []types.DeleteMessageBatchRequestEntry{}

		for _, msg := range messages {
			if err := r.consumer.Consume(msg); err == nil {

				entries = append(entries, types.DeleteMessageBatchRequestEntry{
					Id:            msg.MessageId,
					ReceiptHandle: msg.ReceiptHandle,
				})
			}
		}

		if len(entries) > 0 {
			msgOut <- entries
		}
	}
}

func deleteMessage(ctx context.Context, r *SQSReceive, c <-chan []types.DeleteMessageBatchRequestEntry) {

	for entries := range c {

		gQinput := &sqs.DeleteMessageBatchInput{
			QueueUrl: &r.option.queueUrl,
			Entries:  entries,
		}

		if _, err := r.client.DeleteMessageBatch(context.TODO(), gQinput); err != nil {
			logger.Error().Err(err).
				Interface("batch", entries).
				Msg("batch was not deleted")
		}
	}
}
