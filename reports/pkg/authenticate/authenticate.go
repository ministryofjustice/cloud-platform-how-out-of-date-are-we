// package authenticate creates a clientset for Kubernetes authentication.
package authenticate

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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

	// create the clientset
	clientset, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return
}
