package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
)

var (
	bucket           = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	kubecfgBucketKey = flag.String("kubecfgBucketKey", os.Getenv("KUBECONFIG_S3_KEY"), "Name of kubeconfig file in S3 bucket")
	ctxLive          = flag.String("contextLive", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	ctxManager       = flag.String("contextManager", "manager.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	hoodawApiKey     = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint   = flag.String("hoodawEndpoint", "/helm_whatup", "Endpoint to send the data to")
	hoodawHost       = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	region           = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")
	kubeCfgPath      = flag.String("kubeCfgPath", os.Getenv("KUBECONFIG"), "Path of the kube config file")

	endPoint = *hoodawHost + *hoodawEndpoint
)

type helmNamespace struct {
	Namespace string // key of JSON object "namespace" from helm list
}

type helmRelease struct {
	Name             string `json:"name"`              // key of JSON object "name" from helm whatup
	Namespace        string `json:"namespace"`         // key of JSON object "namespace" from helm whatup
	InstalledVersion string `json:"installed_version"` // key of JSON object "installed_version" from helm whatup
	LatestVersion    string `json:"latest_version"`    // key of JSON object "latest_version" from helm whatup
	Chart            string `json:"chart"`             // key of JSON object "chart" from helm whatup
}

type resourceMap map[string]interface{}

func main() {

	contexts := []string{*ctxLive, *ctxManager}

	var clusters []resourceMap
	// Output the results of `helm whatup` as JSON, for each production cluster
	for _, ctx := range contexts {
		err := authenticate.SwitchContextFromS3Bucket(*bucket, *kubecfgBucketKey, *region, ctx, *kubeCfgPath)
		if err != nil {
			log.Fatalln("error in switching context", err)
		}

		helmListJson, err := executeHelmList()
		if err != nil {
			log.Fatalln("error in executing helm list", err)
		}

		namespaces, err := getNamespaces(helmListJson)
		if err != nil {
			log.Fatalln("error in getting namespaces", err)
		}

		releases, err := getHelmReleasesInNamespaces(namespaces)
		if err != nil {
			log.Fatalln("error in executing helm whatup", err)
		}

		cluster := resourceMap{
			"name": strings.Split(ctx, ".")[0],
			"apps": releases,
		}
		clusters = append(clusters, cluster)
	}

	jsonToPost, err := BuildJsonMap(clusters)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to hoowdaw api
	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// executeHelmList execute helm list all namespaces amd return output as string
func executeHelmList() (string, error) {
	cmd := exec.Command("helm", "list", "--all-namespaces", "-m", "1000", "-o", "json")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return "", err
	}
	return out.String(), nil
}

// getNamespaces takes json output and get list of namespaces
func getNamespaces(helmListJson string) ([]string, error) {

	var namespaces []helmNamespace
	json.Unmarshal([]byte(helmListJson), &namespaces)

	var nsList []string
	for ns := range namespaces {
		nsList = append(nsList, namespaces[ns].Namespace)
	}
	return deduplicateList(nsList), nil
}

// getHelmReleasesInNamespaces takes namespace name and return helm release struct
func getHelmReleasesInNamespaces(namespaces []string) ([]helmRelease, error) {
	var releases []helmRelease
	for _, ns := range namespaces {
		release, err := helmReleasesInNamespace(ns)
		if err != nil {
			log.Fatalln(err.Error())
		}
		releases = append(releases, release...)
	}
	return releases, nil
}

// deduplicateList will take a slice of strings and return a deduplicated version.
func deduplicateList(s []string) (list []string) {
	keys := make(map[string]bool)

	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return
}

// helmReleasesInNamespace takes namespace name and execute helm whatup to return releases
func helmReleasesInNamespace(namespace string) ([]helmRelease, error) {
	cmd := exec.Command("helm", "whatup", "--namespace", namespace, "-o", "json")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return nil, err
	}

	var rel map[string][]helmRelease

	if err := json.Unmarshal(out.Bytes(), &rel); err != nil {
		fmt.Println("Unable to unmarshal JSON: ", err)
		return nil, err
	}

	return rel["releases"], nil

}

// BuildJsonMap takes a slice of maps and return a json encoded map
func BuildJsonMap(clusters []resourceMap) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at": time.Now().Format("2006-01-2 15:4:5 UTC"),
		"clusters":   clusters,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
