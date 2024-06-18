package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	utils "github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
	vpc "orphaned_resources/orphaned_resources/vpc"
)

type OrphanedAwsResrouces struct {
	Vpcs []string `json:"vpcs"`
}

type OrphanedAwsResroucesJson struct {
	orphaned_aws_resources OrphanedAwsResrouces
	updated_at             time.Time
}

const BUCKET_NAME = "cloud-platform-terraform-state"

func getAllTfState() []string {
	client, err := utils.S3Client("eu-west-1")
	if err != nil {
		log.Fatalln(err.Error())
	}

	params := s3.ListObjectsV2Input{
		Bucket:    aws.String(BUCKET_NAME),
		Delimiter: aws.String("terrafom.tfstate"),
		Prefix:    aws.String("aws-accounts/cloud-platform-aws/vpc/"),
	}

	// will return up to 1k results
	output, err := client.ListObjectsV2(context.TODO(), &params)
	if err != nil {
		log.Fatalln(err.Error())
	}

	keys := []string{}
	filenames := []string{}

	for _, item := range output.Contents {
		keys = append(keys, *item.Key)
	}

	for idx, filePath := range keys {
		if strings.Contains(filePath, "tfstate") {
			params := s3.GetObjectInput{
				Bucket: aws.String(BUCKET_NAME),
				Key:    aws.String(filePath),
			}
			obj, err := client.GetObject(context.TODO(), &params)
			if err != nil {
				log.Fatalln(err.Error())
			}

			body, err := io.ReadAll(obj.Body)
			strIdx := fmt.Sprint(idx)

			// TODO: construct more meaningful tfstate file name
			filename := "local_tfstate/" + strIdx + "-terraform.tfstate"

			err = os.WriteFile(filename, body, 0644)
			if err != nil {
				log.Fatalln(err.Error())
			}
			fmt.Printf("writing tfstate to file: " + filename + "\n")

			filenames = append(filenames, filename)
		}
	}

	return filenames
}

func main() {
	tfStateFiles := getAllTfState()

	ec2Client, err := utils.Ec2Client()
	if err != nil {
		log.Fatal(err)
	}

	_, ec2Err := vpc.GetOrphaned(ec2Client, tfStateFiles)
	if ec2Err != nil {
		log.Fatal(ec2Err)
	}

	// construct the json from all the resource functions

	// upload to the hoodaw reports bucket
}
