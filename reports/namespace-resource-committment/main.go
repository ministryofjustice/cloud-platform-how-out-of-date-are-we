package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	authenticate "github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	namespace "github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	// bucket      = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx = flag.String("context", "arn:aws:eks:eu-west-2:754256621582:cluster/live", "Kubernetes context specified in kubeconfig")
	// kubeconfig  = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
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

	// Get pod requests requests of all namespaces of a given cluster
	nsReqMap, nsLimitMap, err := getAllContainerResourceDetails(kclientset)
	if err != nil {
		log.Fatalln("error in getting all pod resources details", err.Error())
	}

	// Get pod usage resources of all namespaces of a given cluster
	nsUsedMap, err := getAllPodMetricsesDetails(mclientset)
	if err != nil {
		log.Fatalln("error in getting all pod metrics details", err.Error())
	}

	limitRangeReqMap, limitRangeLimitMap, err := GetAllLimitRangeDetails(kclientset)
	if err != nil {
		log.Fatalln("error in getting all limit range details", err.Error())
	}

	// write to the csv file
	csvFile, err := os.Create("limitRanges.csv")
	if err != nil {
		log.Fatalln("error in creating csv file", err.Error())
	}
	csvwriter := csv.NewWriter(csvFile)

	csvwriter.Write([]string{"namespace", "pod",
		"reqCPU>limitDefaultReqCPU", "reqMem>limitDefaultReqCPU", "limitCPU>DefaultLimitsCPU", "limitMem>DefaultLimitsMem",
		"requested_cpu", "used_cpu", "limits_cpu",
		"limitRangeReqMap_cpu", "limitRangeLimitMap_cpu",
		"requested_memory", "used_memory", "limits_memory",
		"limitRangeReqMap_memory", "limitRangeLimitMap_memory"})
	for pod, req := range nsReqMap {
		reqCPUGtlimitDefaultReqCPU, reqMemGtlimitDefaultReqCPU, limitCPUGtDefaultLimitsCPU, limitMemGtDefaultLimitsMem := "false", "false", "false", "false"
		if req.CPU > limitRangeReqMap[req.Namespace].CPU {
			reqCPUGtlimitDefaultReqCPU = "true"
		}
		if req.Memory > limitRangeReqMap[req.Namespace].Memory {
			reqMemGtlimitDefaultReqCPU = "true"
		}
		if limitRangeLimitMap[req.Namespace].CPU > limitRangeReqMap[req.Namespace].CPU {
			limitCPUGtDefaultLimitsCPU = "true"
		}
		if limitRangeLimitMap[req.Namespace].Memory > limitRangeReqMap[req.Namespace].Memory {
			limitMemGtDefaultLimitsMem = "true"
		}

		csvwriter.Write([]string{req.Namespace, pod,
			reqCPUGtlimitDefaultReqCPU, reqMemGtlimitDefaultReqCPU, limitCPUGtDefaultLimitsCPU, limitMemGtDefaultLimitsMem,
			fmt.Sprintf("%f", req.CPU),
			fmt.Sprintf("%f", nsUsedMap[pod].CPU),
			fmt.Sprintf("%f", nsLimitMap[pod].CPU),
			fmt.Sprintf("%f", limitRangeReqMap[req.Namespace].CPU),
			fmt.Sprintf("%f", limitRangeLimitMap[req.Namespace].CPU),
			fmt.Sprintf("%f", req.Memory),
			fmt.Sprintf("%f", nsUsedMap[pod].Memory),
			fmt.Sprintf("%f", nsLimitMap[pod].Memory),
			fmt.Sprintf("%f", limitRangeReqMap[req.Namespace].Memory),
			fmt.Sprintf("%f", limitRangeLimitMap[req.Namespace].Memory)})
	}
	csvFile.Close()
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

		for _, container := range pod.Spec.Containers {
			cpuReq, memoryReq := container.Resources.Requests[v1.ResourceCPU], container.Resources.Requests[v1.ResourceMemory]

			r.CPU = float64(cpuReq.MilliValue())
			r.Memory = float64(memoryReq.Value() / 1048576)
			r.Namespace = pod.Namespace
			containerName := pod.Name + "-" + container.Name
			nsReqMap[containerName] = r
			cpuLimits, memoryLimits := container.Resources.Limits[v1.ResourceCPU], container.Resources.Limits[v1.ResourceMemory]

			r.CPU = float64(cpuLimits.MilliValue())
			r.Memory = float64(memoryLimits.Value() / 1048576)
			r.Namespace = pod.Namespace
			nsLimitMap[containerName] = r
		}
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
		r := ContainerResource{}
		for _, container := range podMetrics.Containers {
			cpuUsage, memoryUsage := container.Usage[v1.ResourceCPU], container.Usage[v1.ResourceMemory]
			r.CPU = float64(cpuUsage.MilliValue())
			r.Memory = float64(memoryUsage.Value() / 1048576)
			r.Namespace = podMetrics.Namespace
			containerName := podMetrics.Name + "-" + container.Name
			nsUsedMap[containerName] = r
		}

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

	for _, ns := range namespaces {
		fmt.Println("Getting limit ranges for namespace %s", ns.Name)

		limitRanges, err := kclientset.CoreV1().LimitRanges(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("can't list limitranges from cluster %s", err.Error())
		}
		if len(limitRanges.Items) > 0 {

			limitRange := limitRanges.Items[0]

			// limit := v1.ResourceList{}
			r := ContainerResource{}
			limit := limitRange.Spec.Limits[0]

			cpuLimitReq, memoryLimitReq := limit.DefaultRequest[v1.ResourceCPU], limit.DefaultRequest[v1.ResourceMemory]
			r.CPU = float64(cpuLimitReq.MilliValue())
			r.Memory = float64(memoryLimitReq.Value() / 1048576)
			r.Namespace = ns.Name
			nsReqMap[ns.Name] = r

			cpuLimit, memoryLimit := limit.Default[v1.ResourceCPU], limit.Default[v1.ResourceMemory]

			r.CPU = float64(cpuLimit.MilliValue())
			r.Memory = float64(memoryLimit.Value() / 1048576)
			r.Namespace = ns.Name
			nsLimitMap[ns.Name] = r
		} else {
			fmt.Println("No limit ranges found for namespace %s", ns.Name)
			continue
		}
	}

	return nsReqMap, nsLimitMap, nil

}
