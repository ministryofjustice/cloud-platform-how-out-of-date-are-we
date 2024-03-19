package utils

import (
	"bytes"
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

var ctx = context.TODO()

func awsS3Client() (*s3.Client, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	cfg := s3.NewFromConfig(sdkConfig)

	return cfg, nil
}

func CheckBucketExists(bucket string) (bool, error) {
	client, err := awsS3Client()
	if err != nil {
		return false, err
	}

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				log.Println("Bucket not found")
				exists = false
				err = nil
			default:
				log.Printf("Either bucket %s exists or there was an error\n", bucket)
			}
		}
	} else {
		log.Printf("Bucket %s exists\n", bucket)
	}

	return exists, err
}

func ExportToS3(bucketName, objectKey string, file []byte) error {
	client, err := awsS3Client()
	if err != nil {
		return err
	}

	client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(file),
	})

	return err
}

func ArchiveFile(bucketName, objectKey string) error {
	client, err := awsS3Client()
	if err != nil {
		return err
	}

	// check if the object exists
	_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return err
	} else {
		_, err = client.CopyObject(ctx, &s3.CopyObjectInput{
			Bucket:     aws.String(bucketName),
			CopySource: aws.String(bucketName + "/" + objectKey),
			Key:        aws.String("archive/" + objectKey),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func ImportS3File(bucketName, objectKey string) ([]byte, error) {
	client, err := awsS3Client()
	if err != nil {
		return nil, err
	}

	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Body)

	return buf.Bytes(), nil
}
