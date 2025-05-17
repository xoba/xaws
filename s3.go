package xaws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// re https://docs.aws.amazon.com/AmazonS3/latest/userguide/qfacts.html
const (
	KB           = 1024         // kilobyte
	MB           = KB * KB      // megabyte
	GB           = KB * KB * KB // gigabyte
	TB           = MB * MB      // terabyte
	MinPartSize  = 5 * MB       // except last part, which can be smaller
	MaxPartSize  = 5 * GB
	MaxTotalSize = 5 * TB
)

// MultipartResponse summarizes the multipart upload result.
type MultipartResponse struct {
	BytesUploaded int
	PartsUploaded int
}

// File represents a readable object that can report its size.
type File interface {
	io.Reader
	Stat() (fs.FileInfo, error)
}

// UploadMultipart performs a multipart upload for large objects.
func UploadMultipart(c context.Context, svc *s3.Client, f File, bucket, key string) (*MultipartResponse, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	switch {
	case fi.Size() < MinPartSize:
		if _, err := svc.PutObject(c, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   f,
		}); err != nil {
			return nil, err
		}
		return &MultipartResponse{
			BytesUploaded: int(fi.Size()),
			PartsUploaded: 1,
		}, nil
	case fi.Size() > MaxTotalSize:
		return nil, fmt.Errorf("file too big: %d vs %d", fi.Size(), MaxTotalSize)
	}
	const partSize = MaxPartSize / 5
	if partSize > MaxPartSize {
		return nil, fmt.Errorf("part size too big: %d vs %d", partSize, MaxPartSize)
	}
	m, err := svc.CreateMultipartUpload(c, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	var n int
	var partNumber int32
	var completedParts []types.CompletedPart
	for {
		partNumber++
		partBuffer := make([]byte, partSize)
		bytesRead, err := f.Read(partBuffer)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		n += bytesRead
		resp, err := svc.UploadPart(context.Background(), &s3.UploadPartInput{
			Bucket:     aws.String(bucket),
			Key:        aws.String(key),
			UploadId:   m.UploadId,
			PartNumber: &partNumber,
			Body:       bytes.NewReader(partBuffer[:bytesRead]),
		})
		if err != nil {
			return nil, err
		}
		completedParts = append(completedParts, types.CompletedPart{
			ETag:       resp.ETag,
			PartNumber: ptr(partNumber),
		})
	}
	if _, err = svc.CompleteMultipartUpload(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: m.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}); err != nil {
		return nil, err
	}
	return &MultipartResponse{
		BytesUploaded: n,
		PartsUploaded: len(completedParts),
	}, nil
}
