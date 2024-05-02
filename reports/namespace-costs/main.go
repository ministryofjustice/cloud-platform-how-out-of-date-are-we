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
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
	v1 "k8s.io/api/core/v1"
)

var (
	hoodawBucket = flag.String("howdaw-bucket", os.Getenv("HOODAW_BUCKET"), "AWS S3 bucket for hoodaw json reports")
	bucket       = flag.String("bucket", os.Getenv("KUBECONFIG_S3_BUCKET"), "AWS S3 bucket for kubeconfig")
	ctx          = flag.String("context", "live.cloud-platform.service.justice.gov.uk", "Kubernetes context specified in kubeconfig")
	kubeconfig   = flag.String("kubeconfig", "kubeconfig", "Name of kubeconfig file in S3 bucket")
	region       = flag.String("region", os.Getenv("AWS_REGION"), "AWS Region")
)

const SHARED_COSTS string = "SHARED_COSTS"

// Annual cost of the Cloud Platform team is £866,100.
// This is based on the FTE of the team, all at Senior WebOps Engineer level.
// This is then converted to USD, divided by 12, to get a monthly cost, and rounded up to the nearest $1000.
// This was last updated on 30/03/2023 using FX rate £1 = $1.24.
const MONTHLY_TEAM_COST = 90_000

const DAYS_TOGET_DATA int = 30

// resourceMap is used to store both string:string and string:map[string]interface{} key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a interface of key value
// pairs with key namespace name and values as another map with keys 'breakdown' and 'total'.
type resourceMap map[string]interface{}

// costs is a map which has the namespace name as key and the value a map
// of resource names as key and costs as value
type costs struct {
	costPerNamespace map[string]map[string]float64
}

func main() {
	flag.Parse()

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

	//create a new costs object
	c := &costs{
		costPerNamespace: map[string]map[string]float64{},
	}

	// get Cost and Usage data from aws cost explorer api
	awsCostUsageData, err := getAwsCostAndUsageData()
	if err != nil {
		log.Fatalln(err.Error())
	}

	// create the resources map for namespaces which are listed in the cluster
	// This is needed later to update shared costs for namespaces which doesnot have any aws resources
	for _, ns := range namespaces {
		resources := make(map[string]float64)
		c.costPerNamespace[ns.Name] = resources
	}

	// update the costs per namespace in a map for all aws resources from CostUsage data
	err = c.updatecostsByNamespace(awsCostUsageData)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// add shared aws resources costs i.e resources which doesnot have namespace tags but global
	// resources to the CP account e.g ec2 instances, elasticsearch
	c.addSharedCosts()

	// add shared CP team costs
	c.addSharedTeamCosts()

	// build the resources Map for all namespaces with the format required by HOODAW frontend
	namespacesMap := c.buildCostsResourceMap(namespaces)

	// build the final jsonMap
	jsonToPost, err := BuildJsonMap(namespacesMap)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to S3
	client, err := utils.S3Client()
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

// getAwsCostAndUsageData get the data from aws cost explorer api and build a slice of [date,resourcename,namespacename,cost]
func getAwsCostAndUsageData() ([][]string, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	svc := costexplorer.NewFromConfig(cfg)
	now, monthBefore := timeNow(DAYS_TOGET_DATA)

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
		return nil, err
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

// timeNow will take the number of days as input and return the current month and the month past 30 days
func timeNow(x int) (string, string) {
	dt := time.Now()
	now := dt.Format("2006-01-02")
	month := dt.AddDate(0, 0, -x).Format("2006-01-02")
	return now, month
}

// updatecostsByNamespace get the aws CostUsageData and update the costPerNamespace
// with resources and map per namespace
func (c *costs) updatecostsByNamespace(awsCostUsageData [][]string) error {

	for _, col := range awsCostUsageData {
		cost, err := strconv.ParseFloat(col[3], 64)
		if err != nil {
			fmt.Println(err)
			return err
		}

		c.addResource(col[2], col[1], cost)

	}
	return nil
}

// addSharedCosts get the value of shared costs for each namespace, delete the shared_costs key and
// and assign the shared_costs per namespace
func (c *costs) addSharedCosts() error {

	costsPerNs := c.getSharedCosts()
	delete(c.costPerNamespace, SHARED_COSTS)
	c.addSharedPerNamespace(costsPerNs)
	return nil

}

// getSharedCosts calculates the shared costs by adding
// all the costs of global resources needed for the Platform and
// divide it by number of namespaces in the cluster
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

// addSharedPerNamespace get the shared cost and assign the shared_costs per namespace
func (c *costs) addSharedPerNamespace(costsPerNs float64) {

	for _, v := range c.costPerNamespace {
		v["Shared AWS Costs"] = costsPerNs
	}

}

// add shared team costs per namespace
func (c *costs) addSharedTeamCosts() error {

	nKeys := len(c.costPerNamespace)
	perNsSharedCPCosts := MONTHLY_TEAM_COST / float64(nKeys)
	roundedCPCost := math.Round(perNsSharedCPCosts*100) / 100

	for _, v := range c.costPerNamespace {
		v["Shared CP Team Costs"] = roundedCPCost
	}

	return nil

}

// buildCostsResourceMap build the resources Map for all namespaces
// with the format required by HOODAW frontend
func (c *costs) buildCostsResourceMap(nsList []v1.Namespace) resourceMap {

	namespaces := make(map[string]interface{}, 0)

	for _, ns := range nsList {
		breakdown := c.costPerNamespace[ns.Name]

		var total float64 = 0
		m := breakdown
		for _, val := range m {
			total += val
		}

		total = math.Round(total*100) / 100
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

// hasResource get the namespace name and resource name and checks if it has value in costPerNamespace
func (c *costs) hasResource(ns, resource string) float64 {
	return c.costPerNamespace[ns][resource]
}
