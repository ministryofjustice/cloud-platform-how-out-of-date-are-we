package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/ingress"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
	networkingv1 "k8s.io/api/networking/v1"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes ingress resources that contain old live1-domain dns address names.
type resourceMap map[string]interface{}

var (
	hoodawBucket  = flag.String("hoodaw-bucket", os.Getenv("HOODAW_BUCKET"), "AWS S3 bucket for hoodaw json reports")
	bucket        = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx           = flag.String("context", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	kubeconfig    = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region        = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")
	liveOneDomain = "live-1.cloud-platform.service.justice.gov.uk"
)

func main() {
	flag.Parse()

	// Gain access to a Kubernetes cluster using a config file stored in an S3 bucket.
	clientset, err := authenticate.CreateClientFromS3Bucket(*bucket, *kubeconfig, *region, *ctx)
	if err != nil {
		err := fmt.Errorf("unable to authenticate to the cluster")
		fmt.Println(err.Error())
		return
	}

	// Get all ingress resources
	domainSearch, err := ingress.GetAllIngressesFromCluster(clientset)
	if err != nil {
		err := fmt.Errorf("unable to return Ingress List from the cluster: %s", err)
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Searching for live-1 domains in ingress resources completed")

	// Find all ingress resources with the live-1-domain name
	ingress := liveOneDomainSearch(domainSearch)

	// Build the json map
	jsonToPost, err := buildJsonMap(ingress)
	if err != nil {
		err := fmt.Errorf("unable to build json map: %s", err)
		fmt.Println(err.Error())
		return
	}

	client, err := utils.S3Client("eu-west-2")
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

	exportErr := utils.ExportToS3(client, *hoodawBucket, "live_one_domains.json", jsonToPost)
	if exportErr != nil {
		log.Fatalln(exportErr.Error())
	}
}

// Live1DomainSearch searches list created by GetAllIngresses for all ingress resources and returns a
// list of namespace, ingress resources and the hosts that are still using the live1-domain name
func liveOneDomainSearch(domainSearch *networkingv1.IngressList) []map[string]string {
	// s contains a slice of maps, each map will be iterated over when placed in a dashboard.
	var s []map[string]string

	for _, domain := range domainSearch.Items {
		for _, rule := range domain.Spec.Rules {
			if strings.Contains(rule.Host, liveOneDomain) {
				ingress := map[string]string{
					"namespace": domain.Namespace,
					"ingress":   domain.Name,
					"hostname":  rule.Host,
					"CreatedAt": domain.CreationTimestamp.Format("2006-01-2 15:4:5 UTC"),
				}
				s = append(s, ingress)
			}
		}
	}
	return s
}

// BuildJsonMap takes a slice of maps and return a json encoded map
func buildJsonMap(ingress []map[string]string) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at":       time.Now().Format("2006-01-2 15:4:5 UTC"),
		"live_one_domains": ingress,
	}

	jsonStr, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
