package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	configFile     = flag.String("configFile", os.Getenv("KUBECONFIG_S3_KEY"), "Name of kubeconfig file in S3 bucket")
	hoodawEndpoint = "/ingress-weighting"
)

func main() {
	flag.Parse()

	buff := &aws.WriteAtBuffer{}
	downloader := s3manager.NewDownloader(session.New(&aws.Config{
		Region: aws.String("eu-west-2"),
	}))

	numBytes, err := downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(*bucket),
		Key:    aws.String(*configFile),
	})

	if err != nil {
		log.Fatalln(err.Error())
	}
	if numBytes < 1 {
		log.Fatalln("The file downloaded is incorrect.")
	}

	data := buff.Bytes()

	// use the current context in kubeconfig
	config, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		log.Println(err.Error())
	}

	clientConfig, err := config.ClientConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		log.Println(err.Error())
	}

	ingress, err := clientset.NetworkingV1beta1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
	}

	// For each ingress resource, check
	m := make(map[string]string)
	for _, i := range ingress.Items {
		if _, ok := i.Annotations["external-dns.alpha.kubernetes.io/aws-weight"]; !ok {
			for _, v := range i.Spec.TLS {
				m[i.GetNamespace()] = i.GetName() + "/" + v.Hosts[0]
			}
		}
	}

	jsonStr, err := json.Marshal(m)
	fmt.Println(string(jsonStr))

	req, err := http.NewRequest("POST", *hoodawHost+*&hoodawEndpoint, nil)

	req.Header.Add("X-API-KEY", *hoodawApiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error on POST response. \n", err)
	}

	defer resp.Body.Close()
}
