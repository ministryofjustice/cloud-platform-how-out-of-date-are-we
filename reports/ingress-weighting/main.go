package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	hoodawApiKey    = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawHost      = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	awsAccessId     = flag.String("accessId", os.Getenv("KUBECONFIG_AWS_ACCESS_KEY_ID"), "Access ID of the AWS account")
	awsAccessSecret = flag.String("accessSecret", os.Getenv("KUBECONFIG_AWS_SECRET_ACCESS_KEY"), "Secret key of the AWS account")
	bucket          = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	configFile      = flag.String("configFile", os.Getenv("KUBECONFIG_S3_KEY"), "Name of kubeconfig file in S3 bucket")
	abs             = flag.String("abs", "/Users/jasonbirchall/Documents/workarea/cloud-platform-how-out-of-date-are-we/reports/ingress-weighting/live-1-only", "Name of kubeconfig file in S3 bucket")
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	// buff := &aws.WriteAtBuffer{}
	// downloader := s3manager.NewDownloader(session.New(&aws.Config{
	// 	Region: aws.String("eu-west-2"),
	// }))

	// numBytes, err := downloader.Download(buff, &s3.GetObjectInput{
	// 	Bucket: aws.String(*bucket),
	// 	Key:    aws.String(*configFile),
	// })
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// data := buff.Bytes()

	// if numBytes < 1 {
	// 	log.Fatalln("Unable to fetch token from s3.")
	// }
	file, err := os.Create(*configFile)
	if err != nil {
		log.Fatalln("Unable to open file %q, %v", configFile, err)
	}

	defer file.Close()

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2")},
	)

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(*bucket),
			Key:    aws.String(*configFile),
		})
	if err != nil {
		log.Fatalln("Unable to download item %q, %v", *configFile, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	flag.Parse()

	absPath, err := filepath.Abs(*configFile)
	if err != nil {
		log.Println("Unable to get working directory.")
	}
	fmt.Println(absPath)

	// loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	// loadingRules.ExplicitPath = absPath
	// configOverides := &clientcmd.ConfigOverrides{
	// 	ClusterDefaults: clientcmd.ClusterDefaults,
	// 	CurrentContext:  "live-1.cloud-platform.service.justice.gov.uk",
	// }

	// config, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverides).ClientConfig()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalln("can't find token", err.Error())
	}
	// fmt.Println(config)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err.Error())
	}

	ingress, err := clientset.NetworkingV1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println("Unable to get ingress", err.Error())
	}

	// For each ingress resource, check
	m := make(map[string]v1.IngressTLS)
	for _, i := range ingress.Items {
		if _, ok := i.Annotations["external-dns.alpha.kubernetes.io/aws-weight"]; !ok {
			m[i.GetNamespace()+"/"+i.GetName()] = i.Spec.TLS[0]
		}
	}

	jsonStr, err := json.Marshal(m)
	fmt.Println(string(jsonStr))
	// config, err := clientcmd.NewClientConfigFromBytes(data)
	// // config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	log.Println(err.Error())
	// }

	// clientConfig, err := config.ClientConfig()
	// if err != nil {
	// 	log.Fatalln("nope", err)
	// }
	// // fmt.Println(clientConfig)

	// // create the clientset
	// clientset, err := kubernetes.NewForConfig(clientConfig)
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	// fmt.Println(clientset)

}
