package subnets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type Value struct {
	Value []string `json:"value"`
}

type SubnetIdsTfState struct {
	ExternalSubnetIds Value `json:"external_subnets_ids"`
	InternalSubnetIds Value `json:"internal_subnets_ids"`
}

type Instances struct {
	Attributes map[string]any `json:"attributes,omitempty"`
}

type Resource struct {
	Instances []Instances `json:"instances"`
}

type SubnetsTfState struct {
	Outputs   SubnetIdsTfState `json:"outputs"`
	Resources []Resource       `json:"resources"`
}

type OrphanedSubnet struct {
	Cluster  string
	SubnetId string
}

func getFromTf(tfStateFiles []string) ([]string, error) {
	subnetIds := []string{}

	for _, file := range tfStateFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var subnetsTfState SubnetsTfState

		unmarshalErr := json.Unmarshal(data, &subnetsTfState)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}

		// outputs -> external_subnets_ids || outputs -> internal_subnets_ids
		if len(subnetsTfState.Outputs.ExternalSubnetIds.Value) > 0 || len(subnetsTfState.Outputs.InternalSubnetIds.Value) > 0 {
			subnetIds = append(subnetIds, subnetsTfState.Outputs.ExternalSubnetIds.Value...)
			subnetIds = append(subnetIds, subnetsTfState.Outputs.InternalSubnetIds.Value...)
		}

		if len(subnetsTfState.Resources) > 0 {
			// loop through resources -> instances -> attributes -> outputs -> value
			for _, resource := range subnetsTfState.Resources {
				for _, instance := range resource.Instances {
					if _, ok := instance.Attributes["value"]; ok {
						externalIds, externalOk := instance.Attributes["value"].(map[string]any)
						internalIds, internalOk := instance.Attributes["value"].(map[string]any)
						if externalOk && internalOk {
							subnetIds = append(subnetIds, externalIds["external_subnets_ids"].([]string)...)
							subnetIds = append(subnetIds, internalIds["internal_subnets_ids"].([]string)...)

						}
					}
				}
			}
		}
	}

	slices.Sort(subnetIds)

	return slices.Compact(subnetIds), nil
}

func GetOrphaned(ec2Client *ec2.Client, tfStateFiles []string) ([]OrphanedSubnet, error) {
	DEFAULT_SUBNET_IDS := []string{"subnet-4178f728", "subnet-cdf6e980", "subnet-a069a0da"}
	orphanedSubnets := []OrphanedSubnet{}
	subnetIds, tfStateErr := getFromTf(tfStateFiles)

	subnetIds = append(subnetIds, DEFAULT_SUBNET_IDS...)

	if tfStateErr != nil {
		return nil, tfStateErr
	}

	awsSubnets, awsSubnetsErr := ec2Client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{})
	if awsSubnetsErr != nil {
		return nil, awsSubnetsErr
	}

	for _, subnet := range awsSubnets.Subnets {
		if !utils.Contains(subnetIds, *subnet.SubnetId) {
			clusterName := ""
			for _, tag := range subnet.Tags {
				if *tag.Key == "Cluster" {
					clusterName = *tag.Value
				}
			}
			orphanedSubnets = append(orphanedSubnets, OrphanedSubnet{clusterName, *subnet.SubnetId})
		}
	}

	fmt.Printf("There are %d Oprhaned Subnets.\n", len(orphanedSubnets))

	if len(orphanedSubnets) > 0 {
		log.Println("Oprhaned Subnet Ids: ", orphanedSubnets)
	}

	return orphanedSubnets, nil
}
