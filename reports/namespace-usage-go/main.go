package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	authenticate "github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	namespace "github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var (
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx            = flag.String("context", "live-1.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
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
	Requested      NamespaceResource
	Used           NamespaceResource
	Hardlimits     NamespaceResource
	ContainerCount int
	Name           string
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

	// Get the list of pods from the cluster which is set in the kclientset
	nsList, err := namespace.GetAllNamespacesFromCluster(kclientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Get the list of pods from the cluster which is set in the kclientset
	podsList, err := namespace.GetAllPodsFromCluster(kclientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	nsReqMap := make(map[string]NamespaceResource, 0)

	containerMap := make(map[string]int, 0)

	// get resource request of each pod and container count
	// and store it in namespaceResource map
	for _, pod := range podsList {
		req, namespace, newCount := GetPodResourceDetails(pod)
		list := nsReqMap[namespace]
		if _, exist := nsReqMap[namespace]; exist {
			list.addNamespaceResource(req)
			nsReqMap[namespace] = list
		} else {
			nsReqMap[namespace] = req
		}
		containerMap[namespace] += newCount
	}

	// Get top pods(resource used) of all namespaces from the cluster which is set in the mclientset
	// podMetricsList := []v1beta1.PodMetrics
	podMetricsList, err := namespace.GetAllPodMetricsesFromCluster(mclientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	nsUsedMap := make(map[string]NamespaceResource, 0)

	for _, podMetrics := range podMetricsList {
		used, namespace := GetPodUsageDetails(podMetrics)
		list := nsUsedMap[namespace]
		if _, exist := nsUsedMap[namespace]; exist {
			list.addNamespaceResource(used)
			nsUsedMap[namespace] = list
		} else {
			nsUsedMap[namespace] = used
		}
	}

	// get namespace quota of namespaces to find hard limits of pods from the cluster
	rsQuotasList, err := namespace.GetAllResourceQuotasFromCluster(kclientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	nsQuotaMap := make(map[string]NamespaceResource, 0)

	for _, rsQuota := range rsQuotasList {
		hardLimits, namespace, err := GetPodHardLimits(rsQuota)
		if err != nil {
			log.Fatalln(err.Error())
		}
		nsQuotaMap[namespace] = hardLimits
	}

	var usageReports []UsageReport

	for _, ns := range nsList {
		var usageReport UsageReport
		usageReport.Name = ns.Name
		usageReport.Requested = nsReqMap[ns.Name]
		usageReport.Used = nsUsedMap[ns.Name]
		usageReport.Hardlimits = nsQuotaMap[ns.Name]
		usageReport.ContainerCount = containerMap[ns.Name]
		usageReports = append(usageReports, usageReport)
	}

	for ns, rep := range usageReports {
		fmt.Println("ns", ns, "value", rep)
	}

}

// GetPodResourceDetails takes a Pod of type v1.Pod and collect
// all resources summed up for all containers of the pod and return the result
func GetPodResourceDetails(pod v1.Pod) (r NamespaceResource, namespace string, containerCount int) {
	reqs, _ := corev1.ResourceList{}, corev1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
		containerCount++
	}
	cpuReq, memoryReq := reqs[corev1.ResourceCPU], reqs[corev1.ResourceMemory]

	r.CPU = float64(cpuReq.MilliValue())
	r.Memory = float64(memoryReq.Value() / 1048576)
	namespace = pod.Namespace
	return
}

// GetPodResourceDetails takes a Pod of type v1.Pod and collect
// all resources summed up for all containers of the pod and return the result
func GetPodUsageDetails(PodMetrics v1beta1.PodMetrics) (u NamespaceResource, namespace string) {

	usage := corev1.ResourceList{}
	for _, container := range PodMetrics.Containers {
		addResourceList(usage, container.Usage)
	}
	cpuUsage, memoryUsage := usage[corev1.ResourceCPU], usage[corev1.ResourceMemory]
	u.CPU = float64(cpuUsage.MilliValue())
	u.Memory = float64(memoryUsage.Value() / 1048576)
	namespace = PodMetrics.Namespace
	return
}

// GetPodResourceDetails takes a Pod of type v1.Pod and collect
// all resources summed up for all containers of the pod and return the result
func GetPodHardLimits(resourceQuota corev1.ResourceQuota) (h NamespaceResource, namespace string, err error) {
	hardLimits := resourceQuota.Status.Hard["pods"].DeepCopy()
	h.Pods, err = strconv.Atoi(hardLimits.String())
	namespace = resourceQuota.Namespace
	if err != nil {
		return NamespaceResource{}, "", err
	}
	return
}

func (list *NamespaceResource) addNamespaceResource(new NamespaceResource) {
	list.CPU = list.CPU + new.CPU
	list.Memory = list.Memory + new.Memory
	list.Pods++
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
