package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
	v1 "k8s.io/api/core/v1"
)

var (
	bucket         = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx            = flag.String("context", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/costs_by_namespace", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	kubeconfig     = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region         = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")

	endPoint = *hoodawHost + *hoodawEndpoint
)

const SHARED_COSTS string = "SHARED_COSTS"

// Annual cost of the Cloud Platform team is £1,260,000
// This is the monthly cost in USD
const MONTHLY_TEAM_COST = 136_000
const SHARED_CP_COSTS string = "Shared CP Team Costs"

type costs struct {
	costPerNamespace map[string]map[string]float64
}

func main() {
	flag.Parse()

	c := &costs{
		costPerNamespace: map[string]map[string]float64{},
	}

	awsCostUsageData, err := GetAwsCostAndUsageData()
	if err != nil {
		log.Fatalln(err.Error())
	}
	// Get all namespaces from cluster

	// Gain access to a Kubernetes cluster using a config file stored in an S3 bucket.
	clientset, err := authenticate.CreateClientFromS3Bucket(*bucket, *kubeconfig, *region, *ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Get the list of namespaces from the cluster which is set in the clientset
	namespaces, err := namespace.GetAllNamespacesFromCluster(clientset)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, ns := range namespaces {
		resources := make(map[string]float64)
		c.costPerNamespace[ns.Name] = resources
	}

	err = c.updatecostsByNamespace(awsCostUsageData)
	if err != nil {
		log.Fatalln(err.Error())
	}

	c.addSharedCosts()

	// add shared team costs
	c.addSharedTeamCosts()

	namespacesMap := c.buildCostsResourceMap(namespaces)

	jsonToPost, err := BuildJsonMap(namespacesMap)
	if err != nil {
		log.Fatalln(err.Error())
	}

	//fmt.Println(namespacesMap)
	// // Post json to hoowdaw api
	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}

	//

}

//func (c *costs) costPerNamespace() map[string]map[string]float64 { return c.costPerNamespace }

func GetAwsCostAndUsageData() ([][]string, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// handle error
	}
	svc := costexplorer.NewFromConfig(cfg)
	now, monthBefore := timeNow(31)

	param := &costexplorer.GetCostAndUsageInput{
		Granularity: ceTypes.GranularityMonthly,
		TimePeriod: &ceTypes.DateInterval{
			Start: aws.String(monthBefore),
			End:   aws.String(now),
		},
		Metrics: []string{"BlendedCost"},
		GroupBy: []ceTypes.GroupDefinition{
			{
				Type: ceTypes.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
			{
				Type: ceTypes.GroupDefinitionTypeTag,
				Key:  aws.String("namespace"),
			},
		},
	}

	GetCostAndUsageOutput, err := svc.GetCostAndUsage(context.TODO(), param)
	if err != nil {
		fmt.Println(err)
	}

	var resultsCosts [][]string
	for _, results := range GetCostAndUsageOutput.ResultsByTime {
		startDate := *results.TimePeriod.Start
		for _, groups := range results.Groups {
			for _, metrics := range groups.Metrics {
				tag_value := strings.Split(groups.Keys[1], "$")
				if tag_value[1] == "" {
					tag_value[1] = SHARED_COSTS
				}
				info := []string{startDate, groups.Keys[0], tag_value[1], *metrics.Amount}

				resultsCosts = append(resultsCosts, info)

			}
		}
	}
	return resultsCosts, nil
}

func timeNow(x int) (string, string) {
	dt := time.Now()
	now := dt.Format("2006-01-02")
	month := dt.AddDate(0, 0, -x).Format("2006-01-02")
	return now, month
}

func (c *costs) updatecostsByNamespace(awsCostUsageData [][]string) error {

	for _, col := range awsCostUsageData {
		cost, err := strconv.ParseFloat(col[3], 64)
		if err != nil {
			fmt.Println(err)
			return err
		}

		c.addResource(col[2], col[1], cost)

	}

	// for k, v := range costsPerNamespaceMap {
	// 	fmt.Println("key[%s] value[%s]\n", k, v)
	// }
	return nil
}

// get the value of shared costs for each namespace, delete the shared_costs key and
// and assign the shared_costs per namespace
func (c *costs) addSharedCosts() error {

	costsPerNs := c.getSharedCosts()
	delete(c.costPerNamespace, SHARED_COSTS)
	c.addSharedPerNamespace(costsPerNs)
	return nil

}

func (c *costs) getSharedCosts() float64 {
	nKeys := len(c.costPerNamespace)

	sharedCosts := c.costPerNamespace[SHARED_COSTS]
	var totalCost float64
	for _, v := range sharedCosts {
		totalCost += v
	}
	// calculate per namespace cost taking away the shared_costs key
	perNsSharedCosts := totalCost / float64(nKeys-1)
	return math.Round(perNsSharedCosts*100) / 100
}

func (c *costs) addSharedPerNamespace(costsPerNs float64) {

	for _, v := range c.costPerNamespace {
		v[SHARED_COSTS] = costsPerNs
	}

}

// add shared team costs per namespace
func (c *costs) addSharedTeamCosts() error {

	nKeys := len(c.costPerNamespace)
	perNsSharedCPCosts := MONTHLY_TEAM_COST / float64(nKeys)
	roundedCPCost := math.Round(perNsSharedCPCosts*100) / 100

	for _, v := range c.costPerNamespace {
		v[SHARED_CP_COSTS] = roundedCPCost
	}

	return nil

}

type resourceMap map[string]interface{}

// add shared team costs per namespace
func (c *costs) buildCostsResourceMap(nsList []v1.Namespace) resourceMap {

	namespaces := make(map[string]interface{}, 0)

	for _, ns := range nsList {
		breakdown := c.costPerNamespace[ns.Name]

		var total float64 = 0
		m := breakdown
		for _, val := range m {
			total += val
		}

		namespaces[ns.Name] = resourceMap{
			"breakdown": breakdown,
			"total":     total,
		}

	}

	return namespaces

}

// BuildJsonMap takes a slice of maps and return a json encoded map
func BuildJsonMap(namespaceMap resourceMap) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]interface{}.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at": time.Now().Format("2006-01-2 15:4:5 UTC"),
		"namespace":  namespaceMap,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}

func (c *costs) addResource(ns, resource string, cost float64) {
	resources := c.costPerNamespace[ns]

	if resources == nil {
		resources = make(map[string]float64)
		c.costPerNamespace[ns] = resources
		resources[resource] = cost
	} else {

		curCost := c.hasResource(ns, resource)
		if curCost == 0 {
			resources[resource] = curCost
		}
		curCost = cost + curCost
		resources[resource] = math.Round(curCost*100) / 100
	}

}

func (c *costs) hasResource(ns, resource string) float64 {
	return c.costPerNamespace[ns][resource]
}
