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
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
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
	CPU       float64
	Memory    float64
	Pods      int
	Namespace string
}
type UsageReport struct {
	Requested  NamespaceResource
	Used       NamespaceResource
	Hardlimits NamespaceResource
	Namespace  string
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

	//nsReqMap := make(map[string]NamespaceResource, 0)

	var nsReq []NamespaceResource

	// get resource request and limits of each pod and store it in namespace map
	for _, pod := range podsList {
		req := GetPodResourceDetails(pod)
		nsReq = append(nsReq, req)
	}

	// count of containers spec.containers.count per namespace

	// resource_used
	// Get top pods of all namespaces and map it with pod map - resource_used

	//podMetricsList := []v1beta1.PodMetrics
	// Get the list of namespaces from the cluster which is set in the clientset
	podMetricsList, err := namespace.GetAllPodMetricsesFromCluster(mclientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var nsUsed []NamespaceResource
	// get resource request and limits of each pod and store it in namespace map
	for _, podMetrics := range podMetricsList {
		used := GetPodUsageDetails(podMetrics)
		nsUsed = append(nsUsed, used)

	}
	// get namespace quota to find hard limits of pods

}

// GetPodResourceDetails takes a Pod of type v1.Pod and collect
// all resources summed up for all containers of the pod and return the result
func GetPodResourceDetails(pod v1.Pod) (r NamespaceResource) {
	reqs, _ := corev1.ResourceList{}, corev1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
	}
	cpuReq, memoryReq := reqs[corev1.ResourceCPU], reqs[corev1.ResourceMemory]

	r.CPU = float64(cpuReq.MilliValue())
	r.Memory = float64(memoryReq.Value() / 1048576)
	r.Namespace = pod.Namespace
	return
}

// GetPodResourceDetails takes a Pod of type v1.Pod and collect
// all resources summed up for all containers of the pod and return the result
func GetPodUsageDetails(PodMetrics v1beta1.PodMetrics) (u NamespaceResource) {

	usage := corev1.ResourceList{}
	for _, container := range PodMetrics.Containers {
		addResourceList(usage, container.Usage)
	}
	cpuUsage, memoryUsage := usage[corev1.ResourceCPU], usage[corev1.ResourceMemory]
	u.CPU = float64(cpuUsage.MilliValue())
	u.Memory = float64(memoryUsage.Value() / 1048576)
	u.Namespace = PodMetrics.Namespace
	return
}

// addResourceList adds the resources in newList to list
func addResourceList(list, new corev1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}
