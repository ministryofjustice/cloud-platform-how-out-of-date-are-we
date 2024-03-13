package hoodaw

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	ctx = context.Background()
)

func awsS3Client() *s3.Client {
	cfg := s3.NewFromConfig(aws.Config{
		Region: "eu-west-2",
	})

	return cfg
}

func ExportToS3(data []byte, bucket string, key string) {
	client := awsS3Client()

	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully uploaded to S3")
}
