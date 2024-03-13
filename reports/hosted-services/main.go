package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/ingress"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/aws"
	networkingv1 "k8s.io/api/networking/v1"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes namespaces and hostnames that don't contain the annotation contained within
// the variable 'annotation'.
type resourceMap map[string]interface{}

var (
	hoodawBucket   = flag.String("howdaw-bucket", os.Getenv("HOODAW_BUCKET"), "AWS S3 bucket for hoodaw json reports")
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx            = flag.String("context", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/hosted_services", "Endpoint to send the data to")
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
		log.Fatalln(err.Error())
	}

	// Get the list of namespaces from the cluster which is set in the clientset
	namespaces, err := namespace.GetAllNamespacesFromCluster(clientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	//make namespace map
	nsDetailsMap := make(map[string]namespace.Namespace, 0)

	// get required details of each namespace and store it in namespace map
	for _, ns := range namespaces {
		namespaceDetails := GetNamespaceDetails(ns)
		nsDetailsMap[namespaceDetails.Name] = namespaceDetails
	}

	// Get all ingress resources
	ingressList, err := ingress.GetAllIngressesFromCluster(clientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Build the ingresses Map
	ingressesMap := BuildIngressesMap(ingressList.Items)

	// Add the ingress slice to the namespace Details map for all namespaces
	for k, v := range nsDetailsMap {
		if _, exist := ingressesMap[v.Name]; exist {
			v.DomainNames = ingressesMap[v.Name]
		}
		nsDetailsMap[k] = v
	}

	jsonToPost, err := BuildJsonMap(nsDetailsMap)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to S3
	msg, err := aws.ExportToS3(jsonToPost, *hoodawBucket, "hosted_services.json")
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println(msg)

}

// GetNamespaceDetails takes a Namespace of type v1.namespace and stores the required annotations
// and labels into a struct and return the result
func GetNamespaceDetails(ns v1.Namespace) namespace.Namespace {
	return namespace.Namespace{
		Name:             ns.Name,
		Application:      ns.Annotations["cloud-platform.justice.gov.uk/application"],
		BusinessUnit:     ns.Annotations["cloud-platform.justice.gov.uk/business-unit"],
		DeploymentType:   ns.Labels["cloud-platform.justice.gov.uk/environment-name"],
		GithubURL:        ns.Annotations["cloud-platform.justice.gov.uk/source-code"],
		TeamName:         ns.Annotations["cloud-platform.justice.gov.uk/team-name"],
		TeamSlackChannel: ns.Annotations["cloud-platform.justice.gov.uk/slack-channel"],
		DomainNames:      []string{},
	}
}

// BuildIngressesMap takes the Ingress list and return a map with key as namespace and value
// with slices of string containing hosts urls
func BuildIngressesMap(ingressItems []networkingv1.Ingress) map[string][]string {

	ingressMap := make(map[string][]string, 0)

	for _, i := range ingressItems {

		for _, v := range i.Spec.TLS {
			if len(v.Hosts) > 0 {
				ingressMap[i.Namespace] = append(ingressMap[i.Namespace], v.Hosts[0])
			}
		}
	}
	return ingressMap
}

// BuildJsonMap takes a map with namespce key and namespace struct as value, sort the map, flatten to a
// slice and return a json encoded map
func BuildJsonMap(hostedservices map[string]namespace.Namespace) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.

	// sort the keys by ascending order
	keys := make([]string, 0, len(hostedservices))
	for key := range hostedservices {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// flatten the map to a slice which is expected by the HOODAW API
	flattenMap := make([]namespace.Namespace, 0)

	for _, k := range keys {
		flattenMap = append(flattenMap, hostedservices[k])
	}

	jsonMap := resourceMap{
		"updated_at":        time.Now().Format("2006-01-2 15:4:5 UTC"),
		"namespace_details": flattenMap,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
