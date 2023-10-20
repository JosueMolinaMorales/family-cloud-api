package config

import (
	"context"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config/log"
	"github.com/JosueMolinaMorales/family-cloud-api/pkg/error"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// AwsDriver is the interface for the aws driver
type AwsDriver interface {
	ListObjects(ctx context.Context, params *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, *error.RequestError)
}

// NewAwsDriver creates a new aws driver
func NewAwsDriver(logger log.Logger) AwsDriver {
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
	logger log.Logger
}

func (a *awsDriver) ListObjects(ctx context.Context, params *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, *error.RequestError) {
	res, err := a.client.ListObjectsV2(ctx, params)
	if err != nil {
		// check if error is a context error
		if err.Error() == context.Canceled.Error() {
			return nil, error.NewRequestError(err, error.BadRequestError, "request timed out", a.logger)
		}
		return nil, error.NewRequestError(err, error.InternalServerError, "failed to list objects", a.logger)
	}

	return res, nil
}
