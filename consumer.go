package aksqs

import "github.com/aws/aws-sdk-go-v2/service/sqs/types"

type CustomConsumer struct {
}

func NewCustomConsumer() *CustomConsumer {

	return &CustomConsumer{}
}

func (c *CustomConsumer) Consume(msg types.Message) error {
	logger.Info().Interface("mt", msg.Body).Msg("success")
	return nil
}
