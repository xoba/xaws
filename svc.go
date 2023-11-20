package xaws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

func NewSesV2() (*sesv2.Client, error) {
	c, err := newConfig()
	if err != nil {
		return nil, err
	}
	return sesv2.NewFromConfig(c), nil
}

const MaxEmailBytesSESV2 = 40 * 1024 * 1024

func newConfig(optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return config.LoadDefaultConfig(context.Background(), optFns...)
}
