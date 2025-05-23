package xaws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

// NewSesV2 returns a configured SESv2 client.
func NewSesV2() (*sesv2.Client, error) {
	c, err := newConfig()
	if err != nil {
		return nil, err
	}
	return sesv2.NewFromConfig(c), nil
}

// eliminates gratuitous warnings, thanks to https://tsak.dev/posts/aws-sdk-suppress-checksum-warning/
// e.g., SDK 2025/02/17 16:36:44 WARN Response has no supported checksum. Not validating response payload.
// NewS3 returns a configured S3 client.
func NewS3() (*s3.Client, error) {
	c, err := newConfig()
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(c, func(o *s3.Options) {
		o.DisableLogOutputChecksumValidationSkipped = true
	}), nil
}

// NewKMS returns a configured KMS client.
func NewKMS() (*kms.Client, error) {
	c, err := newConfig()
	if err != nil {
		return nil, err
	}
	return kms.NewFromConfig(c), nil
}

// NewSFN returns a configured Step Functions client.
func NewSFN() (*sfn.Client, error) {
	c, err := newConfig()
	if err != nil {
		return nil, err
	}
	return sfn.NewFromConfig(c), nil
}

const MaxEmailBytesSESV2 = 40 * 1024 * 1024

// newConfig loads AWS configuration with optional overrides.
func newConfig(optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return config.LoadDefaultConfig(context.Background(), optFns...)
}
