package vpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	utils "github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type Value struct {
	Value string `json:"value"`
}

type NetworkIdTfState struct {
	NetworkId Value `json:"network_id"`
}

type VpcTfState struct {
	Outputs NetworkIdTfState `json:"outputs"`
}

type OrhpanedVpc struct {
	Cluster string
	VpcId   string
}

func getFromTf(tfStateFiles []string) ([]string, error) {
	vpcIds := []string{}

	for _, file := range tfStateFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var vpcState VpcTfState

		unmarshalErr := json.Unmarshal(data, &vpcState)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}

		if vpcState.Outputs.NetworkId.Value != "" {
			vpcIds = append(vpcIds, vpcState.Outputs.NetworkId.Value)
		}
	}

	return vpcIds, nil
}

func GetOrphaned(ec2Client *ec2.Client, tfStateFiles []string) ([]OrhpanedVpc, error) {
	DEFAULT_VPC_ID := "vpc-057ac86d"
	orphanedVpcs := []OrhpanedVpc{}
	vpcIds, tfStateErr := getFromTf(tfStateFiles)

	vpcIds = append(vpcIds, DEFAULT_VPC_ID)

	if tfStateErr != nil {
		return nil, tfStateErr
	}

	awsVpcs, awsVpcErr := ec2Client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})

	if awsVpcErr != nil {
		return nil, awsVpcErr
	}

	for _, vpc := range awsVpcs.Vpcs {
		if !utils.Contains(vpcIds, *vpc.VpcId) {
			clusterName := ""
			for _, tag := range vpc.Tags {
				if *tag.Key == "Cluster" {
					clusterName = *tag.Value
				}
			}
			orphanedVpcs = append(orphanedVpcs, OrhpanedVpc{clusterName, *vpc.VpcId})
		}
	}

	fmt.Printf("There are %d Oprhaned VPCs.\n", len(orphanedVpcs))

	if len(orphanedVpcs) > 0 {
		log.Println("Oprhaned VPC Ids:", orphanedVpcs)
	}

	return orphanedVpcs, nil
}
