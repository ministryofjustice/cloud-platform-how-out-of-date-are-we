package utils

import (
	"bytes"
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

var ctx = context.TODO()

// S3AssumeRole returns an S3 client with assumed role
func S3AssumeRole(roleArn, roleSessionName string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	stsClient := sts.NewFromConfig(cfg)

	creds := stscreds.NewAssumeRoleProvider(stsClient, roleArn, func(o *stscreds.AssumeRoleOptions) {
		o.RoleSessionName = roleSessionName
	})

	cfg.Credentials = aws.NewCredentialsCache(creds)

	return s3.NewFromConfig(cfg), nil
}

// S3Client returns an S3 client
func S3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

// CheckBucketExists checks if a bucket exists
func CheckBucketExists(client *s3.Client, bucket string) (bool, error) {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
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

// ExportToS3 uploads a file to S3
func ExportToS3(client *s3.Client, bucketName, objectKey string, file []byte) error {
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(file),
	})

	return err
}

// ArchiveFile copies the object to the archive folder in the same bucket
func ArchiveFile(client *s3.Client, bucketName, objectKey string) error {
	// check if the object exists
	_, err := client.HeadObject(ctx, &s3.HeadObjectInput{
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

// ImportS3File downloads a file from S3 and returns the file content and the last modified timestamp
func ImportS3File(client *s3.Client, bucketName, objectKey string) ([]byte, string, error) {
	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, "", err
	}

	// read the file content
	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Body)

	// get the last modified timestamp
	fileTimeStamp := output.LastModified.String()

	return buf.Bytes(), fileTimeStamp, nil
}
