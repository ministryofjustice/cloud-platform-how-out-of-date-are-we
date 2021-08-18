package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	// "github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/authenticate"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes namespaces and hostnames that don't contain the annotation contained within
// the variable 'annotation'.
type resourceMap map[string]interface{}

var (
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	configFile     = flag.String("configFile", os.Getenv("KUBECONFIG_S3_KEY"), "Name of kubeconfig file in S3 bucket")
	annotation     = flag.String("annotation", "external-dns.alpha.kubernetes.io/aws-weight", "String of the annotation to check")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/ingress_weighting", "Endpoint to send the data to")
	endPoint       = *hoodawHost + *hoodawEndpoint
)

func main() {
	flag.Parse()

	// Gain access to a Kubernetes cluster using a config file stored in an S3 bucket.
	// clientset, err := authenticate.FromS3Bucket(*bucket, *configFile)
	clientset, err := FromS3Bucket(*bucket, *configFile)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Grab all ingress resources without required annotation
	jsonToPost, err := IngressWithoutAnnotation(clientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to hoowdaw api
	err = postToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// FromS3Bucket accepts two strings, a bucket and a configFile. The bucket string should
// contain the name of an S3 bucket that contains a kubeconfig file. The configFile string
// should contain the kubeconfig file name held within the bucket. Both of these values are
// defined by flags passed to main and default to an environment variable. The function returns
// a Kubernetes clientset and an error, if there is one. The clientset uses the current context
// value in the kubeconfig file, so this must be set beforehand.
func FromS3Bucket(bucket, configFile string) (clientset *kubernetes.Clientset, err error) {
	buff := &aws.WriteAtBuffer{}
	downloader := s3manager.NewDownloader(session.New(&aws.Config{
		Region: aws.String("eu-west-2"),
	}))

	numBytes, err := downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(configFile),
	})

	if err != nil {
		return nil, err
	}
	if numBytes < 1 {
		return nil, errors.New("The file downloaded is incorrect.")
	}

	data := buff.Bytes()

	// use the current context in kubeconfig
	config, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		return nil, err
	}

	clientConfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return
}

// postToApi takes a slice of bytes as an argument and attempts to POST it to the REST API
// provided by HOODAW. The slice of bytes should contain json using the guidelines outlined
// by HOODAW i.e. the first entry in the key value pair should contain a string:string, which consists
// of a string and the time POSTed.
func postToApi(jsonToPost []byte, hoodawApiKey, endPoint *string) error {
	req, err := http.NewRequest("POST", *endPoint, bytes.NewBuffer(jsonToPost))
	if err != nil {
		return err
	}

	req.Header.Add("X-API-KEY", *hoodawApiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
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
		if _, ok := i.Annotations[*annotation]; !ok {
			for _, v := range i.Spec.TLS {
				m := make(map[string]string)
				m["namespace"] = i.GetName()
				m["hostname"] = v.Hosts[0]
				s = append(s, m)
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
