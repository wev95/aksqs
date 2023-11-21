package aksqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var client *sqs.Client

type ClientOption struct {
	Profile string
	Host    string
	Region  string
}

func New(ctx context.Context, opt *ClientOption) (*sqs.Client, error) {

	var (
		cfg aws.Config
		err error
	)

	if (opt.Profile == "dev" || opt.Profile == "") && opt != nil {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           opt.Host,
				SigningRegion: opt.Region,
			}, nil
		})

		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(opt.Region), config.WithEndpointResolverWithOptions(customResolver))

	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(opt.Region))
	}

	client = sqs.NewFromConfig(cfg)
	return client, err
}
