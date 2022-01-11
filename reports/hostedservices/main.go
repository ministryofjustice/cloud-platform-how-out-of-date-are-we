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
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes namespaces and hostnames that don't contain the annotation contained within
// the variable 'annotation'.
type resourceMap map[string]interface{}

// Namespace describes a Cloud Platform namespace object.
type namespace struct {
	Application      string
	BusinessUnit     string
	DeploymentType   string
	Cluster          string
	DomainNames      []string
	GithubURL        string
	Name             string
	RbacTeam         []string
	TeamName         string
	TeamSlackChannel string
}

// // AllNamespaces contains the list of namespaces of type Namespace.
// type allNamespaces struct {
// 	Namespaces []namespace
// }

var (
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

	namespaces, _ := GetNamespaces(clientset)

	//make namespace map
	nsDetailsMap := make(map[string]namespace, 0)

	for _, ns := range namespaces {

		namespaceDetails := GetNamespaceDetails(ns)
		nsDetailsMap[namespaceDetails.Name] = namespaceDetails
	}

	// Get all ingress resources
	ingressList, err := GetAllIngresses(clientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ingressMap := make(map[string][]string, 0)

	for _, i := range ingressList.Items {

		for _, v := range i.Spec.TLS {
			if len(v.Hosts) > 0 {
				ingressMap[i.Namespace] = append(ingressMap[i.Namespace], v.Hosts[0])
			}
		}
	}

	for k, v := range nsDetailsMap {
		v.DomainNames = ingressMap[v.Name]
		nsDetailsMap[k] = v
	}

	jsonToPost, err := BuildJsonMap(nsDetailsMap)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to hoowdaw api
	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}

}

// GetNamespaces takes a Kubernetes clientset and returns all namespaces with type *v1beta1.IngressList and an error.

func GetNamespaceDetails(ns v1.Namespace) namespace {

	var namespaceDetails namespace

	namespaceDetails.Name = ns.Name
	namespaceDetails.Application = ns.Annotations["cloud-platform.justice.gov.uk/application"]
	cluster := strings.SplitN(*ctx, ".", 2)
	namespaceDetails.Cluster = cluster[0]
	namespaceDetails.BusinessUnit = ns.Annotations["cloud-platform.justice.gov.uk/business-unit"]
	namespaceDetails.DeploymentType = ns.Labels["cloud-platform.justice.gov.uk/environment-name"]
	namespaceDetails.GithubURL = ns.Annotations["cloud-platform.justice.gov.uk/source-code"]
	namespaceDetails.TeamName = ns.Annotations["cloud-platform.justice.gov.uk/team-name"]
	namespaceDetails.TeamSlackChannel = ns.Annotations["cloud-platform.justice.gov.uk/slack-channel"]
	namespaceDetails.DomainNames = []string{"[ ]"}
	return namespaceDetails
}

// GetNamespaces takes a Kubernetes clientset and returns all namespaces with type *v1beta1.IngressList and an error.

func GetNamespaces(clientSet *kubernetes.Clientset) ([]v1.Namespace, error) {

	namespaces, err := clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalln("Can'list namespaces", err.Error())
		return nil, err
	}

	return namespaces.Items, nil
}

// GetAllIngresses takes a Kubernetes clientset and returns all ingress with type *v1beta1.IngressList and an error.
func GetAllIngresses(clientset *kubernetes.Clientset) (*v1beta1.IngressList, error) {
	ingressList, err := clientset.NetworkingV1beta1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return ingressList, nil
}

// BuildJsonMap takes a slice of maps and return a json encoded map
func BuildJsonMap(hostedservices map[string]namespace) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.

	flattenMap := make([]namespace, 0)
	for _, value := range hostedservices {
		flattenMap = append(flattenMap, value)
	}

	// fmt.Println(flattenMap)
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
