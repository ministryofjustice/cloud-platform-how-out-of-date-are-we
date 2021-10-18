package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes namespaces and hostnames that don't contain the annotation contained within
// the variable 'annotation'.
type resourceMap map[string]interface{}

var (
	awsweightAnnotation     = flag.String("awsweightAnnotation", "external-dns.alpha.kubernetes.io/aws-weight", "String of the aws weight annotation to check")
	setIdentifierAnnotation = flag.String("setIdentifierAnnotation", "external-dns.alpha.kubernetes.io/set-identifier", "String of the set-identiifer annotation to check")
	bucket                  = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx                     = flag.String("context", "live-1.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	hoodawApiKey            = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint          = flag.String("hoodawEndpoint", "/ingress_weighting", "Endpoint to send the data to")
	hoodawHost              = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	kubeconfig              = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region                  = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")

	endPoint = *hoodawHost + *hoodawEndpoint
)

func main() {
	flag.Parse()

	// Gain access to a Kubernetes cluster using a config file stored in an S3 bucket.
	err := authenticate.KubeConfigFromS3Bucket(*bucket, *kubeconfig, *region)
	if err != nil {
		log.Fatalln(err.Error())
	}

	clientset, err := authenticate.KubeClientFromConfig("~/.kube/config", *ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Get all ingress resources
	ingressList, err := GetAllIngresses(clientset)
	if err != nil {
		log.Fatalln(err.Error())
	}
	// Find all ingress resources without the required external-dns annotations
	ingressesWithoutAnnotation, err := IngressWithoutAnnotation(ingressList)
	if err != nil {
		log.Fatalln(err.Error())
	}

	jsonToPost, err := BuildJsonMap(ingressesWithoutAnnotation)
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

// IngressWithoutAnnotation takes a list of ingress in  type *v1beta1.IngressList and returns a slice of byte and an error,
// if there is one. Due to the requirement of the API, we have to sculpt the []byte data a very specific way.
func IngressWithoutAnnotation(ingressList *v1beta1.IngressList) ([]map[string]string, error) {
	// s contains a slice of maps, each map will be iterated over when placed in a dashboard.
	s := make([]map[string]string, 0)

	// For each ingress resource, check if it contains the required annotation. If not,
	// loop over the slice of TLS hostnames contained in the resource, create a map, add
	// the namespace and hostname values and add it to a slice of maps.
	for _, i := range ingressList.Items {
		_, exists := i.Annotations[*awsweightAnnotation]
		setIdentifierInIngress, identifierExists := i.Annotations[*setIdentifierAnnotation]
		if !exists || (!identifierExists || setIdentifierInIngress != strings.Join([]string{i.GetName(), i.Namespace, "blue"}, "-")) {
			for _, v := range i.Spec.TLS {
				if len(v.Hosts) > 0 {
					m := make(map[string]string)
					m["namespace"] = i.Namespace
					m["resource"] = i.GetName()
					m["hostname"] = v.Hosts[0]
					s = append(s, m)
				}
			}
		}
	}
	return s, nil
}

// BuildJsonMap takes a slice of maps and return a json encoded map
func BuildJsonMap(ingressesWithoutAnnotation []map[string]string) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at":        time.Now().Format("2006-01-2 15:4:5 UTC"),
		"weighting_ingress": ingressesWithoutAnnotation,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
