package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	client "github.com/ministryofjustice/cloud-platform-cli/pkg/client"

	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cluster "github.com/ministryofjustice/cloud-platform-cli/pkg/cluster"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

var (
	hoodawBucket   = flag.String("howdaw-bucket", os.Getenv("HOODAW_BUCKET"), "AWS S3 bucket for hoodaw json reports")
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/helm_whatup", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	region         = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")
	kubeCfgPath    = flag.String("kubeCfgPath", os.Getenv("KUBECONFIG"), "Path of the kube config file")

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
	contexts := []string{"live", "manager"}

	var clusters []resourceMap
	// Output the results of `helm whatup` as JSON, for each production cluster
	for _, ctx := range contexts {

		creds, err := getCredentials(*region)
		if err != nil {
			log.Fatalln("failed to get aws creds: %w", err)
		}

		clientset, err := cluster.AuthToCluster(ctx, creds.Eks, *kubeCfgPath, creds.Profile)
		if err != nil {
			log.Fatalln("failed to auth to cluster: %w", err)
		}

		namespaces, err := getCPNamespaces(clientset)
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

		fmt.Println(cluster)
		clusters = append(clusters, cluster)
	}

	jsonToPost, err := BuildJsonMap(clusters)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to S3
	client, err := utils.S3Client()
	if err != nil {
		log.Fatalln(err.Error())
	}

	b, err := utils.CheckBucketExists(client, *hoodawBucket)
	if err != nil {
		log.Fatalln(err.Error())
	}

	if !b {
		log.Fatalf("Bucket %s does not exist\n", *hoodawBucket)
	}

	utils.ExportToS3(client, *hoodawBucket, "hosted_services.json", jsonToPost)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to hoowdaw api
	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func getCredentials(awsRegion string) (*client.AwsCredentials, error) {
	creds, err := client.NewAwsCreds(awsRegion)
	if err != nil {
		return nil, err
	}

	return creds, nil
}

// getCPNamespaces to get list of namespaces with "cloud-platform-out-of-hours-alert" annotation set to true
func getCPNamespaces(clientset kubernetes.Interface) ([]string, error) {
	var nsList []string
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []string{}, err
	}

	for _, ns := range namespaces.Items {
		// fetch namespaces which has specific annotations
		if _, ok := ns.Annotations["cloud-platform-out-of-hours-alert"]; ok {
			nsList = append(nsList, ns.Name)
		}

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

	fmt.Println("Executing helm whatup in namespace %s", namespace, "and got the following releases: ", rel["releases"])

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
