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
	setIdentifierAnnotation = flag.String("setIdentifierAnnotation", "external-dns.alpha.kubernetes.io/set-idenitifier", "String of the set-identiifer annotation to check")
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
	clientset, err := authenticate.FromS3Bucket(*bucket, *kubeconfig, *ctx, *region)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Grab all ingress resources without required annotation
	jsonToPost, err := IngressWithoutAnnotation(clientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to hoowdaw api
	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// IngressWithoutAnnotation takes a Kubernetes clientset and returns a slice of byte and an error,
// if there is one. Due to the requirement of the API, we have to sculpt the []byte data a very specific way.
func IngressWithoutAnnotation(clientset *kubernetes.Clientset) ([]byte, error) {
	ingress, err := clientset.NetworkingV1beta1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// s contains a slice of maps, each map will be iterated over when placed in a dashboard.
	s := make([]map[string]string, 0)

	// For each ingress resource, check if it contains the required annotation. If not,
	// loop over the slice of TLS hostnames contained in the resource, create a map, add
	// the namespace and hostname values and add it to a slice of maps.
	for _, i := range ingress.Items {
		//	if _, exists := i.Annotations[*annotation]; !exists {

		_, exists := i.Annotations[*awsweightAnnotation]
		setIdentifierInIngress, _ := i.Annotations[*setIdentifierAnnotation]
		if !exists || setIdentifierInIngress != strings.Join([]string{i.Namespace, i.GetName(), "blue"}, "-") {
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

	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at":        time.Now().Format("2006-01-2 15:4:5 UTC"),
		"weighting_ingress": s,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
