package subnets

import (
	"encoding/json"
	"fmt"
	"os"
)

type Value struct {
	Value []string `json:"value"`
}

type SubnetIdsTfState struct {
	ExternalSubnetIds Value `json:"external_subnets_ids"`
	InternalSubnetIds Value `json:"internal_subnets_ids"`
}

type ValueObj struct {
	ExternalSubnetIds []string `json:"external_subnets_ids"`
	InternalSubnetIds []string `json:"internal_subnets_ids"`
}

type Outputs struct {
	Value ValueObj `json:"value"`
}

type Instances struct {
	Attributes Outputs `json:"attributes"`
}

type Resource struct {
	Instances []Instances `json:"instances"`
}

type SubnetsTfState struct {
	Outputs   SubnetIdsTfState `json:"outputs"`
	Resources []Resource       `json:"resources"`
}

func GetFromTf(tfStateFiles []string) ([]string, error) {
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
			fmt.Printf("looping.... %+v\n", subnetsTfState)
			subnetIds = append(subnetIds, subnetsTfState.Outputs.ExternalSubnetIds.Value...)
			subnetIds = append(subnetIds, subnetsTfState.Outputs.InternalSubnetIds.Value...)

		}

		if len(subnetsTfState.Resources) > 0 {
			for _, resource := range subnetsTfState.Resources {
				for _, instance := range resource.Instances {
					subnetIds = append(subnetIds, instance.Attributes.Value.ExternalSubnetIds...)
					subnetIds = append(subnetIds, instance.Attributes.Value.InternalSubnetIds...)
				}
			}
		}
	}

	// loop through resources -> instances -> attributes -> outputs -> value

	// external = external_subnets_ids
	// internal = internal_subnets_ids

	// sort and delete duplicates

	return subnetIds, nil
}

func GetOrphaned() {}
