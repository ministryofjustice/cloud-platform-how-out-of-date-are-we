package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"pkg/mod/github.com/pkg/errors@v0.9.1"

	authenticate "github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	namespace "github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	bucket      = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx         = flag.String("context", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	kubeconfig  = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region      = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")
	kubeCfgPath = flag.String("kubeCfgPath", os.Getenv("KUBECONFIG"), "Path of the kube config file")
)

// NamespaceResource has the type of resource info
// being collected per namespace by this report
type ContainerResource struct {
	CPU       float64
	Memory    float64
	Namespace string
}

// UsageReport is used to store details of requested resources, used resources,
// hardlimits of pods and number of containers per namespace. This is the set of
// data output from this package.
type UsageReport struct {
	Requested      ContainerResource
	Used           ContainerResource
	Limits         ContainerResource
	DefaultRequest ContainerResource
	DefaultLimit   ContainerResource
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
		log.Fatalln("error in creating clientset", err.Error()
	}

	// Get the clientset object to access cluster metrics
	mclientset, err := authenticate.CreateMetricsClientFromConfigFile(*kubeCfgPath, *ctx)
	if err != nil {
		log.Fatalln("error in creating metrics clientset", err.Error())
	}

	// Get pod requests requests of all namespaces of a given cluster
	nsReqMap, nsLimitMap, err := getAllContainerResourceDetails(kclientset)
	if err != nil {
		log.Fatalln("error in getting all pod resources details", err.Error())
	}

	fmt.Println(nsReqMap)
	fmt.Println(nsLimitMap)

	// Get pod usage resources of all namespaces of a given cluster
	nsUsedMap, err := getAllPodMetricsesDetails(mclientset)
	if err != nil {
		log.Fatalln("error in getting all pod metrics details", err.Error())
	}

	fmt.Println(nsUsedMap)

	limitRangeReqMap, limitRangeLimitMap, err := GetAllLimitRangeDetails(kclientset)
	if err != nil {
		log.Fatalln("error in getting all limit range details", err.Error())
	}

	fmt.Println(limitRangeReqMap)
	fmt.Println(limitRangeLimitMap)

}

// getAllPodResourceDetails takes a clientset and return Pod resource details
// of all namespaces in a map and map of container count of all namespaces
func getAllContainerResourceDetails(kclientset kubernetes.Interface) (
	map[string]ContainerResource, map[string]ContainerResource, error) {

	// Get the list of pods from the cluster which is set in the kclientset
	podsList, err := namespace.GetAllPodsFromCluster(kclientset)
	if err != nil {
		return nil, nil, fmt.Errorf("error in getting all pods from cluster %s", err.Error())
	}
	nsReqMap := make(map[string]ContainerResource, 0)
	nsLimitMap := make(map[string]ContainerResource, 0)
	for _, pod := range podsList {
		r := ContainerResource{}

		reqs := v1.ResourceList{}
		limits := v1.ResourceList{}
		for _, container := range pod.Spec.Containers {
			addResourceList(reqs, container.Resources.Requests)
			addResourceList(limits, container.Resources.Limits)
		}
		cpuReq, memoryReq := reqs[v1.ResourceCPU], reqs[v1.ResourceMemory]

		r.CPU = float64(cpuReq.MilliValue())
		r.Memory = float64(memoryReq.Value() / 1048576)
		r.Namespace = pod.Namespace
		nsReqMap[pod.Name] = r
		cpuLimits, memoryLimits := limits[v1.ResourceCPU], limits[v1.ResourceMemory]

		r.CPU = float64(cpuLimits.MilliValue())
		r.Memory = float64(memoryLimits.Value() / 1048576)
		r.Namespace = pod.Namespace
		nsLimitMap[pod.Name] = r
	}
	return nsReqMap, nsLimitMap, nil
}

// getAllPodMetricsesDetails takes a clientset and return Pod usage details from the
// pod metrics of all namespaces
func getAllPodMetricsesDetails(mclientset versioned.Interface) (
	map[string]ContainerResource, error) {

	// Get top pods(resource used) of all namespaces from the cluster which is set in the mclientset
	podMetricsList, err := namespace.GetAllPodMetricsesFromCluster(mclientset)
	if err != nil {
		return nil, fmt.Errorf("error in getting all pods metrics from cluster %s", err.Error())
	}

	nsUsedMap := make(map[string]ContainerResource, 0)

	for _, podMetrics := range podMetricsList {
		usage := v1.ResourceList{}
		r := ContainerResource{}
		for _, container := range podMetrics.Containers {
			addResourceList(usage, container.Usage)
		}
		cpuUsage, memoryUsage := usage[v1.ResourceCPU], usage[v1.ResourceMemory]
		r.CPU = float64(cpuUsage.MilliValue())
		r.Memory = float64(memoryUsage.Value() / 1048576)
		r.Namespace = podMetrics.Namespace
		nsUsedMap[podMetrics.Name] = r

	}
	return nsUsedMap, nil

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

func GetAllLimitRangeDetails(kclientset *kubernetes.Clientset) (map[string]ContainerResource, map[string]ContainerResource, error) {
	namespaces, err := namespace.GetAllNamespacesFromCluster(kclientset)
	if err != nil {
		log.Fatalln("error in getting all namespaces from cluster", err.Error())
	}
	nsReqMap := make(map[string]ContainerResource, 0)
	nsLimitMap := make(map[string]ContainerResource, 0)

	for _, namespace := range namespaces {
		limitrange, err := GetNamespaceLimitRanges(kclientset, namespace)
		if err != nil {
			log.Fatalln("error in getting all limit ranges from cluster", err.Error())
		}
		for _, limit := range limitRange.Spec.Limits {
			if limit.Type == v1.LimitTypeContainer {
				r.CPU = float64(limit.DefaultRequest[v1.ResourceCPU].MilliValue())
				r.Memory = float64(limit.DefaultRequest[v1.ResourceMemory].Value() / 1048576)
				r.Namespace = limitRange.Namespace
				nsReqMap[limitRange.Namespace] = r

				r.CPU = float64(limit.DefaultLimit[v1.ResourceCPU].MilliValue())
				r.Memory = float64(limit.DefaultLimit[v1.ResourceMemory].Value() / 1048576)
				r.Namespace = limitRange.Namespace
				nsLimitMap[limitRange.Namespace] = r

			}
		}
	}
	return nsReqMap, nsLimitMap, nil
}

func GetNamespaceLimitRanges(clientset *kubernetes.Clientset, namespace string) (*corev1.LimitRange, error) {
	limitRanges, err := clientset.CoreV1().LimitRanges(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list limit ranges")
	}

	if len(limitRanges.Items) == 0 {
		return nil, nil
	}

	return &limitRanges.Items[0], nil
}
