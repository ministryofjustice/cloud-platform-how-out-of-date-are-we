package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	hoodawBucket = flag.String("howdaw-bucket", os.Getenv("HOODAW_BUCKET"), "AWS S3 bucket for hoodaw json reports")
	bucket       = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx          = flag.String("context", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	kubeconfig   = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region       = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")
	kubeCfgPath  = flag.String("kubeCfgPath", os.Getenv("KUBECONFIG"), "Path of the kube config file")
)

// NamespaceResource has the type of resource info
// being collected per namespace by this report
type NamespaceResource struct {
	CPU    float64
	Memory float64
	Pods   int
}

// UsageReport is used to store details of requested resources, used resources,
// hardlimits of pods and number of containers per namespace. This is the set of
// data output from this package.
type UsageReport struct {
	Requested      NamespaceResource
	Used           NamespaceResource
	Hardlimits     NamespaceResource
	ContainerCount int
	Name           string
}

func main() {
	flag.Parse()

	// Get the kubeconfig file stored in an S3 bucket.
	err := authenticate.KubeConfigFromS3Bucket(*bucket, *kubeconfig, *region, *kubeCfgPath)
	if err != nil {
		log.Fatalln("error in getting the kubeconfig from s3 bucket", err.Error())
	}

	// Get the clientset to access the k8s cluster
	kclientset, err := authenticate.CreateClientFromConfigFile(*kubeCfgPath, *ctx)
	if err != nil {
		log.Fatalln("error in creating clientset", err.Error())
	}

	// Get the clientset object to access cluster metrics
	mclientset, err := authenticate.CreateMetricsClientFromConfigFile(*kubeCfgPath, *ctx)
	if err != nil {
		log.Fatalln("error in creating metrics clientset", err.Error())
	}

	// Get the list of namespaces from the cluster which is set in the kclientset
	nsList, err := namespace.GetAllNamespacesFromCluster(kclientset)
	if err != nil {
		log.Fatalln("error in getting all namespaces from cluster", err.Error())
	}

	// Get pod requests requests and container count of all namespaces of a given cluster
	nsReqMap, containerMap, err := getAllPodResourceDetails(kclientset)
	if err != nil {
		log.Fatalln("error in getting all pod resources details", err.Error())
	}

	// Get pod usage resources of all namespaces of a given cluster
	nsUsedMap, err := getAllPodMetricsesDetails(mclientset)
	if err != nil {
		log.Fatalln("error in getting all pod metrics details", err.Error())
	}

	// Get hard limit of pods of all namespaces of a given cluster
	nsQuotaMap, err := getAllResourceQuotaDetails(kclientset)
	if err != nil {
		log.Fatalln("error in getting all resourcequota details", err.Error())
	}

	var usageReports []UsageReport
	// Build the total usageReport
	for _, ns := range nsList {
		usageReport := UsageReport{
			Name:           ns.Name,
			Requested:      nsReqMap[ns.Name],
			Used:           nsUsedMap[ns.Name],
			Hardlimits:     nsQuotaMap[ns.Name],
			ContainerCount: containerMap[ns.Name],
		}
		usageReports = append(usageReports, usageReport)
	}

	jsonToPost, err := buildJsonMap(usageReports)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to S3
	client, err := utils.S3Client("eu-west-1")
	if err != nil {
		log.Fatalln(err.Error())
	}

	b, err := utils.CheckBucketExists(client, *hoodawBucket)
	if err != nil {
		log.Fatalln(err.Error())
	}

	if !b {
		log.Fatalf("Bucket %s does not exist\n", *hoodawBucket)
	}

	utils.ExportToS3(client, *hoodawBucket, "namespace_usage.json", jsonToPost)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// getAllPodResourceDetails takes a clientset and return Pod resource details
// of all namespaces in a map and map of container count of all namespaces
func getAllPodResourceDetails(kclientset kubernetes.Interface) (
	map[string]NamespaceResource, map[string]int, error,
) {
	// Get the list of pods from the cluster which is set in the kclientset
	podsList, err := namespace.GetAllPodsFromCluster(kclientset)
	if err != nil {
		return nil, nil, fmt.Errorf("error in getting all pods from cluster %s", err.Error())
	}

	nsReqMap := make(map[string]NamespaceResource, 0)

	containerMap := make(map[string]int, 0)

	// get resource request of each pod and container count
	// and store it in namespaceResource map

	for _, pod := range podsList {
		req, namespace, newCount := getPodResourceDetails(pod)
		list := nsReqMap[namespace]
		if _, exist := nsReqMap[namespace]; exist {
			list.addNamespaceResource(req)
			nsReqMap[namespace] = list
		} else {
			nsReqMap[namespace] = req
		}
		containerMap[namespace] += newCount
	}

	return nsReqMap, containerMap, nil
}

// getAllPodMetricsesDetails takes a clientset and return Pod usage details from the
// pod metrics of all namespaces
func getAllPodMetricsesDetails(mclientset versioned.Interface) (
	map[string]NamespaceResource, error,
) {
	// Get top pods(resource used) of all namespaces from the cluster which is set in the mclientset
	podMetricsList, err := namespace.GetAllPodMetricsesFromCluster(mclientset)
	if err != nil {
		return nil, fmt.Errorf("error in getting all pods metrics from cluster %s", err.Error())
	}

	nsUsedMap := make(map[string]NamespaceResource, 0)

	for _, podMetrics := range podMetricsList {
		used, namespace := getPodUsageDetails(podMetrics)
		list := nsUsedMap[namespace]
		if _, exist := nsUsedMap[namespace]; exist {
			list.addNamespaceResource(used)
			nsUsedMap[namespace] = list
		} else {
			nsUsedMap[namespace] = used
		}
	}
	return nsUsedMap, nil
}

// getAllPodMetricsesDetails takes a clientset, get resourcequotas of all namespaces from the cluster
// and return the hard limits set for the pods of all namespaces
func getAllResourceQuotaDetails(kclientset kubernetes.Interface) (
	map[string]NamespaceResource, error,
) {
	// get namespace quota of namespaces to find hard limits of pods from the cluster
	rsQuotasList, err := namespace.GetAllResourceQuotasFromCluster(kclientset)
	if err != nil {
		return nil, fmt.Errorf("error in getting all resourcequota from cluster %s", err.Error())
	}

	nsQuotaMap := make(map[string]NamespaceResource, 0)

	for _, rsQuota := range rsQuotasList {
		hardLimits, namespace, err := getPodHardLimits(rsQuota)
		if err != nil {
			log.Fatalln(err.Error())
		}
		nsQuotaMap[namespace] = hardLimits
	}
	return nsQuotaMap, nil
}

// getPodResourceDetails takes a Pod of type v1.Pod and collect
// all resources summed up for all containers of the pod and return the result
func getPodResourceDetails(pod v1.Pod) (r NamespaceResource, namespace string, containerCount int) {
	reqs := v1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
		containerCount++
	}
	cpuReq, memoryReq := reqs[v1.ResourceCPU], reqs[v1.ResourceMemory]

	r.CPU = float64(cpuReq.MilliValue())
	r.Memory = float64(memoryReq.Value() / 1048576)
	namespace = pod.Namespace
	return
}

// getPodResourceDetails takes a Pod of type v1.Pod and collect
// all resources summed up for all containers of the pod and return the result
func getPodUsageDetails(podMetrics v1beta1.PodMetrics) (u NamespaceResource, namespace string) {
	usage := v1.ResourceList{}
	for _, container := range podMetrics.Containers {
		addResourceList(usage, container.Usage)
	}
	cpuUsage, memoryUsage := usage[v1.ResourceCPU], usage[v1.ResourceMemory]
	u.CPU = float64(cpuUsage.MilliValue())
	u.Memory = float64(memoryUsage.Value() / 1048576)
	namespace = podMetrics.Namespace
	return
}

// getPodResourceDetails takes a Pod of type v1.Pod and collect
// all resources summed up for all containers of the pod and return the result
func getPodHardLimits(resourceQuota v1.ResourceQuota) (h NamespaceResource, namespace string, err error) {
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
func addResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}

// buildJsonMap takes a array of usageReport struct and return a json encoded map
func buildJsonMap(usageReports []UsageReport) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := hoodaw.ResourceMap{
		"updated_at": time.Now().Format("2006-01-2 15:4:5 UTC"),
		"data":       usageReports,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
