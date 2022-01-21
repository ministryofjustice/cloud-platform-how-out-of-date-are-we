package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
)

var (
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctxLive        = flag.String("contextLive", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	ctxManager     = flag.String("contextManager", "manager.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	ctxLive_1      = flag.String("contextLive_1", "live-1.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/helm_whatup", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	kubeconfig     = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region         = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")

	endPoint = *hoodawHost + *hoodawEndpoint
)

type helmNamespace struct {
	Namespace string
}

type helmRelease struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	InstalledVersion string `json:"installed_version"`
	LatestVersion    string `json:"latest_version"`
	Chart            string `json:"chart"`
}

type resourceMap map[string]interface{}

func main() {

	// export kube config file path.
	os.Setenv("KUBECONFIG", "/tmp/config")

	helmReleaseLive, err := getHelmReleases("live")
	if err != nil {
		log.Fatalln("error in getting helm releases")
	}

	helmReleaseManager, err := getHelmReleases("manager")
	if err != nil {
		log.Fatalln("error in getting helm releases")
	}

	helmReleaseLive1, err := getHelmReleases("live-1")
	if err != nil {
		log.Fatalln("error in getting helm releases")
	}

	clusters := joinAllHelmReleases(helmReleaseLive, helmReleaseManager, helmReleaseLive1)

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

func getHelmReleases(cluster string) ([]helmRelease, error) {

	switch cluster {
	case "live":
		err := authenticate.SwitchContextFromS3Bucket(*bucket, *kubeconfig, *region, *ctxLive)
		if err != nil {
			return nil, err
		}
	case "live-1":
		err := authenticate.SwitchContextFromS3Bucket(*bucket, *kubeconfig, *region, *ctxLive_1)
		if err != nil {
			return nil, err
		}
	case "manager":
		err := authenticate.SwitchContextFromS3Bucket(*bucket, *kubeconfig, *region, *ctxManager)
		if err != nil {
			return nil, err
		}
	default:
		fmt.Println("No cluster given")
	}

	// Get all helm releases in namespaces

	helmListJson, err := executeHelmList()
	if err != nil {
		return nil, err
	}

	namespaces, err := getNamespaces(helmListJson)
	if err != nil {
		return nil, err
	}

	releases, err := getHelmReleasesInNamespaces(namespaces)
	if err != nil {
		return nil, err
	}
	return releases, nil

}

func executeHelmList() (string, error) {
	cmd := exec.Command("helm", "list", "--all-namespaces", "-o", "json")

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

func getNamespaces(helmListJson string) ([]string, error) {

	var namespaces []helmNamespace
	json.Unmarshal([]byte(helmListJson), &namespaces)

	var nsList []string
	for ns := range namespaces {
		nsList = append(nsList, namespaces[ns].Namespace)
	}
	return deduplicateList(nsList), nil
}

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

	fmt.Printf("%+v", rel["releases"])
	return rel["releases"], nil

}

func joinAllHelmReleases(helmReleaseLive, helmReleaseManger, helmReleaseLive_1 []helmRelease) []resourceMap {
	var clusters []resourceMap

	cluster_live := resourceMap{
		"name": "live",
		"apps": helmReleaseLive,
	}

	cluster_manager := resourceMap{
		"name": "manager",
		"apps": helmReleaseManger,
	}

	cluster_live_1 := resourceMap{
		"name": "live-1",
		"apps": helmReleaseLive_1,
	}

	clusters = append(clusters, cluster_live, cluster_manager, cluster_live_1)

	fmt.Println(clusters)
	return clusters
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
