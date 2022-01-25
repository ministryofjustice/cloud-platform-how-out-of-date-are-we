package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	authenticate "github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	namespace "github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

var (
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx            = flag.String("context", "manager.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/hosted_services", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	kubeconfig     = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region         = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")

	endPoint = *hoodawHost + *hoodawEndpoint
)

type NamespaceResource struct {
	CPURequests    float64
	CPULimits      float64
	MemoryRequests float64
	MemoryLimits   float64
	Pods           int
	Namespace      string
}

func main() {

	flag.Parse()

	// Gain access to a Kubernetes cluster using a config file stored in an S3 bucket.

	configFileLocation := filepath.Join("/", "tmp", "config")
	err := authenticate.KubeConfigFromS3Bucket(*bucket, *kubeconfig, *region)
	if err != nil {
		log.Fatalln(err.Error())
	}

	kclientset, err := authenticate.CreateClientFromConfigFile(configFileLocation, *ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	mclientset, err := authenticate.CreateMetricsClientFromConfigFile(configFileLocation, *ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Get the list of namespaces from the cluster which is set in the clientset
	podsList, err := namespace.GetAllPodsFromCluster(kclientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	nsResMap := make(map[string]NamespaceResource, 0)

	// get resource request and limits of each pod and store it in namespace map
	for _, pod := range podsList {
		nsRes := GetPodResourceDetails(pod)
		nsResMap[nsRes.Namespace] = nsRes
	}

	// count of containers spec.containers.count per namespace

	// resource_used
	// Get top pods of all namespaces and map it with pod map - resource_used

	// Get the list of namespaces from the cluster which is set in the clientset
	podMetricsList, err := namespace.GetAllPodMetricsesFromCluster(mclientset)
	if err != nil {
		log.Fatalln(err.Error())
	}
	// hard_limits
	// ns_quota
	// get namespace quota to find hard limits of pods

}

// GetPodResourceDetails takes a Pod of type v1.Pod and collect
// all resources summed up for all containers of the pod and return the result
func GetPodResourceDetails(pod v1.Pod) (nsRes NamespaceResource) {
	nsRes.Namespace = pod.Namespace
	req, limit := resourcehelper.PodRequestsAndLimits(&pod)
	cpuReq, cpuLimit, memoryReq, memoryLimit := req[corev1.ResourceCPU], limit[corev1.ResourceCPU], req[corev1.ResourceMemory], limit[corev1.ResourceMemory]
	nsRes.CPURequests = float64(cpuReq.MilliValue())
	nsRes.CPULimits = float64(cpuLimit.MilliValue())
	nsRes.MemoryRequests = float64(memoryReq.Value())
	nsRes.MemoryLimits = float64(memoryLimit.Value())
	return
}
