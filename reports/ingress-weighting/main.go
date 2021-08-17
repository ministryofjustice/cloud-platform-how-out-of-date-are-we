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

type resourceMap map[string]interface{}

var (
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	configFile     = flag.String("configFile", os.Getenv("KUBECONFIG_S3_KEY"), "Name of kubeconfig file in S3 bucket")
	annotation     = flag.String("annotation", "external-dns.alpha.kubernetes.io/aws-weight", "String of the annotation to check")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/ingress_weighting", "Endpoint to send the data to")
)

func main() {
	flag.Parse()

	// Authenticate to a Kubernetes cluster
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
	err = postToApi(jsonToPost)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

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

	// create the clientset
	clientset, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return
}

func postToApi(jsonToPost []byte) error {
	req, err := http.NewRequest("POST", *hoodawHost+*hoodawEndpoint, bytes.NewBuffer(jsonToPost))
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

func IngressWithoutAnnotation(clientset *kubernetes.Clientset) ([]byte, error) {
	ingress, err := clientset.NetworkingV1beta1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// s contains a slice of maps, each map will be iterated over when placed in a dashboard.
	s := make([]map[string]string, 0)

	// For each ingress resource, check
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

	postToJson := resourceMap{
		"updated_at":        time.Now().Format("2006-01-2 15:4:5 UTC"),
		"weighting_ingress": s,
	}

	jsonStr, err := json.Marshal(postToJson)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
