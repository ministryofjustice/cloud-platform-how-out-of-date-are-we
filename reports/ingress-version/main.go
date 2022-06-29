package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/ingress"
	"k8s.io/api/networking/v1beta1"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes ingress resources that contain old live1-domain dns address names.
type resourceMap map[string]interface{}

var (
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx            = flag.String("context", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/ingress_version", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	kubeconfig     = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region         = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")

	endPoint = *hoodawHost + *hoodawEndpoint
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
	ingList, err := ingress.GetAllIngressesFromCluster(clientset)
	if err != nil {
		err := fmt.Errorf("unable to return Ingress List from the cluster: %s", err)
		fmt.Println(err.Error())
		return
	}

	ingVer := ingressVersion(ingList)
	//ingressVersion(ingList)

	// Build the json map
	jsonToPost, err := buildJsonMap(ingVer)
	if err != nil {
		err := fmt.Errorf("unable to build json map: %s", err)
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("%s\n", jsonToPost)

	// Post json to hoowdaw api
	//err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	//if err != nil {
	//	err := fmt.Errorf("unable to post data to the 'How out of date are we' API: %s", err)
	//	fmt.Println(err.Error())
	//	return
	//}
}

// Live1DomainSearch searches list created by GetAllIngresses for all ingress resources and returns a
// list of namespace, ingress resources and the hosts that are still using the live1-domain name
func ingressVersion(ingList *v1beta1.IngressList) []map[string]string {
	// s contains a slice of maps, each map will be iterated over when placed in a dashboard.
	s := make([]map[string]string, 0)
	for _, i := range ingList.Items {
		// Create a new map for each ingress resource
		if strings.Contains(i.Namespace, "jacksapp-dev") {
			m := make(map[string]string)
			m["namespace"] = i.Namespace
			m["ingress"] = i.Name
			m["anno"] = i.Annotations["kubectl.kubernetes.io/last-applied-configuration"]
			s = append(s, m)
		}
	}
	return s
}

// BuildJsonMap takes a slice of maps and return a json encoded map
func buildJsonMap(ingVer []map[string]string) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at":      time.Now().Format("2006-01-2 15:4:5 UTC"),
		"ingress_version": ingVer,
	}

	jsonStr, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		return nil, err
	}
	return jsonStr, nil
}
