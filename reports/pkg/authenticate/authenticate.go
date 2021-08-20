// package authenticate creates a clientset for Kubernetes authentication.
package authenticate

import (
	"errors"
	"io/ioutil"

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
func FromS3Bucket(bucket, kubeconfig, ctx, region string) (clientset *kubernetes.Clientset, err error) {
	buff := &aws.WriteAtBuffer{}
	downloader := s3manager.NewDownloader(session.New(&aws.Config{
		Region: aws.String(region),
	}))

	numBytes, err := downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(kubeconfig),
	})

	if err != nil {
		return nil, err
	}
	if numBytes < 1 {
		return nil, errors.New("The file downloaded is incorrect.")
	}

	data := buff.Bytes()
	err = ioutil.WriteFile(kubeconfig, data, 0644)
	if err != nil {
		return nil, err
	}

	client, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: ctx,
		}).ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, _ = kubernetes.NewForConfig(client)
	if err != nil {
		return nil, err
	}

	return
}
