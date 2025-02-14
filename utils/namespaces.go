package utils

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Namespaces struct {
	Namespace map[string]string `json:"namespace"`
}

func GetNamespaces(bucket, namespace string, client *s3.Client) ([]string, error) {

	byteValue, _, err := ImportS3File(client, bucket, "namespace_costs.json")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var namespaces Namespaces
	json.Unmarshal(byteValue, &namespaces)

	var namespaceList []string
	for k, _ := range namespaces.Namespace {
		namespaceList = append(namespaceList, k)
	}

	return namespaceList, nil
}
