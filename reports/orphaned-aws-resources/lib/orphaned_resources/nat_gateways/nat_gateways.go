package natgateways

import (
	"encoding/json"
	"os"
)

type Id struct {
	NatGatewayId string `json:"nat_gateway_id,omitempty"`
}

type Instances struct {
	Attributes Id `json:"attributes,omitempty"`
}

type Resource struct {
	Name      string      `json:"name"`
	Instances []Instances `json:"instances"`
}

type NatGatewayTfState struct {
	Resources []Resource `json:"resources"`
}

type OrphanedNats struct {
	Cluster      string
	NatGatewayId string
}

func getFromTf(tfStateFiles []string) ([]string, error) {
	natGatewayIds := []string{}

	for _, file := range tfStateFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var gatewayState NatGatewayTfState

		unmarshalErr := json.Unmarshal(data, &gatewayState)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}

		// loop through resources and find "name" == "private_nat_gateway" -> instances -> attributes => nat_gateway_id
		if len(gatewayState.Resources) > 0 {
			for _, resource := range gatewayState.Resources {
				if resource.Name == "private_nat_gateway" {
					for _, instance := range resource.Instances {
						natGatewayIds = append(natGatewayIds, instance.Attributes.NatGatewayId)
					}
				}
			}
		}
	}

	return natGatewayIds, nil
}

func GetOrphaned(ec2Client *ec2.Client, tfStateFiles []string) ([]OrphanedNats, error) {
	orphanedNatGateways := []OrphanedNats{}
	// natGatewayIds, tfStateErr := getFromTf(tfStateFiles)

	// if tfStateErr != nil {
	// 	return nil, tfStateErr
	// }

	return orphanedNatGateways, nil
}
