package config

import (
	"context"

	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// AwsDriver is the interface for the aws driver
type AwsDriver interface {
	ListObjects(ctx context.Context, params *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
}

// NewAwsDriver creates a new aws driver
func NewAwsDriver(logger Logger) AwsDriver {
	cfg, err := aws_config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	return &awsDriver{
		client: s3.NewFromConfig(cfg),
	}
}

type awsDriver struct {
	client *s3.Client
	logger Logger
}

func (a *awsDriver) ListObjects(ctx context.Context, params *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	res, err := a.client.ListObjectsV2(ctx, params)
	if err != nil {
		return nil, err
	}

	return res, nil
}
