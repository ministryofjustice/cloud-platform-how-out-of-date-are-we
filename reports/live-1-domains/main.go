package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/ingress"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	hoodawEndpoint = flag.String("hoodawEndpoint", "/live_1_domains", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	kubeconfig     = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region         = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")

	liveOneDomain = "live-1.cloud-platform.service.justice.gov.uk"
	endPoint      = *hoodawHost + *hoodawEndpoint
)

func main() {
	flag.Parse()

	// Gain access to a Kubernetes cluster using a config file stored in an S3 bucket.
	clientset, err := authenticate.CreateClientFromS3Bucket(*bucket, *kubeconfig, *region, *ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Get all ingress resources
	ingressList, err := ingress.GetAllIngressesFromCluster(clientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Find all ingress resources without the required external-dns annotations
	ingressesLive1DoaminSearch, err := Live1DomainSearch(ingressList)
	if err != nil {
		log.Fatalln(err.Error())
	}

	jsonToPost, err := BuildJsonMap(ingressesLive1DoaminSearch)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to hoowdaw api
	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// GetAllIngresses takes a Kubernetes clientset and returns all ingress with type *v1beta1.IngressList and an error.
func GetAllIngresses(clientset *kubernetes.Clientset) (*v1beta1.IngressList, error) {
	ingressList, err := clientset.NetworkingV1beta1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return ingressList, nil
}

// Live1DomainSearch searches all ingress resources and returns a list of namespace, ingress resources and the hosts that are still using the live1-domain name.
// if there is one. Due to the requirement of the API, we have to sculpt the []byte data a very specific way.
func Live1DomainSearch(ingressList *v1beta1.IngressList) ([]map[string]string, error) {
	// s contains a slice of maps, each map will be iterated over when placed in a dashboard.
	s := make([]map[string]string, 0)
	for _, i := range ingressList.Items {

		for _, v := range i.Spec.TLS {
			if strings.Contains(v.Hosts[0], liveOneDomain) {
				m := make(map[string]string)
				m["namespace"] = i.Namespace
				m["ingress"] = i.Name
				m["hostname"] = v.Hosts[0]
				s = append(s, m)
			}
		}
	}
	return s, nil
}

// BuildJsonMap takes a slice of maps and return a json encoded map
func BuildJsonMap(ingressesLive1DoaminSearch []map[string]string) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at":     time.Now().Format("2006-01-2 15:4:5 UTC"),
		"live_1_ingress": ingressesLive1DoaminSearch,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
