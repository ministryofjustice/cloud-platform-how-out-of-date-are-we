package orphanedresources

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	VPC_HOME                   = "https://eu-west-2.console.aws.amazon.com/vpc/home?region=eu-west-2"
	EC2_HOME                   = "https://eu-west-2.console.aws.amazon.com/ec2/v2/home?region=eu-west-2"
	RDS_HOME                   = "https://eu-west-2.console.aws.amazon.com/rds/home?region=eu-west-2"
	KOPS_CLUSTER_INSTANCES_URL = "https://eu-west-2.console.aws.amazon.com/ec2/v2/home?region=eu-west-2#Instances:tag:KubernetesCluster"
	NAT_GATEWAY_URL            = VPC_HOME + "#NatGatewayDetails:natGatewayId="
	INTERNET_GATEWAY_URL       = VPC_HOME + "#InternetGateway:internetGatewayId="
	ROUTE_TABLE_URL            = VPC_HOME + "#RouteTables:search="
	SUBNET_URL                 = VPC_HOME + "#subnets:search="
	VPC_URL                    = VPC_HOME + "#VpcDetails:VpcId="
	DEFAULT_VPC_ID             = "vpc-057ac86d"
	KOPS_INSTANCE_TAG          = "kops.k8s.io/instancegroup"
	KOPS_CLUSTER_TAG           = "KubernetesCluster"
)

var IGNORE_ZONE_NAMES = []string{
	"integrationtest.service.justice.gov.uk",
}

type AwsResources struct {
	s3client      *s3.S3
	ec2client     *ec2.EC2
	route53client *route53.Route53
	rdsclient     *rds.RDS
}

func NewAwsResources(sess *session.Session) *AwsResources {
	return &AwsResources{
		s3client:      s3.New(sess),
		ec2client:     ec2.New(sess),
		route53client: route53.New(sess),
		rdsclient:     rds.New(sess),
	}
}

func (r *AwsResources) KopsClusters() []map[string]interface{} {
	result := []map[string]interface{}{}
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
	}))
	resource := ec2.New(sess)
	instances, err := resource.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error describing instances:", err)
		return result
	}

	h := make(map[string]int)
	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			tags := make([]string, len(instance.Tags))
			for i, tag := range instance.Tags {
				tags[i] = *tag.Key
			}
			if contains(tags, KOPS_INSTANCE_TAG) {
				cluster := getTagValue(instance.Tags, KOPS_CLUSTER_TAG)
				h[cluster]++
			}
		}
	}

	for cluster, instances := range h {
		result = append(result, map[string]interface{}{
			"cluster":   cluster,
			"instances": instances,
			"href":      KOPS_CLUSTER_INSTANCES_URL + "=" + cluster,
		})
	}

	return result
}

func (r *AwsResources) Vpcs() []ResourceTuple {
	result := []ResourceTuple{}
	vpcs, err := r.ec2client.DescribeVpcs(nil)
	if err != nil {
		fmt.Println("Error describing VPCs:", err)
		return result
	}

	for _, vpc := range vpcs.Vpcs {
		if *vpc.VpcId == DEFAULT_VPC_ID {
			continue
		}
		url := VPC_URL + *vpc.VpcId
		result = append(result, ResourceTuple{
			ID:            *vpc.VpcId,
			AwsConsoleURL: url,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

func (r *AwsResources) NatGateways() []ResourceTuple {
	result := []ResourceTuple{}
	for _, vpc := range r.Vpcs() {
		result = append(result, r.natGatewayIdsForVpc(vpc.ID)...)
	}
	return cleanList(result)
}

func (r *AwsResources) Subnets() []ResourceTuple {
	result := []ResourceTuple{}
	for _, vpc := range r.Vpcs() {
		result = append(result, r.subnetIds(vpc.ID)...)
	}
	return cleanList(result)
}

func (r *AwsResources) RouteTables() []ResourceTuple {
	result := []ResourceTuple{}
	for _, subnet := range r.Subnets() {
		result = append(result, r.routeTablesForSubnet(subnet.ID)...)
	}
	return cleanList(result)
}

func (r *AwsResources) RouteTableAssociations() []ResourceTuple {
	result := []ResourceTuple{}
	for _, subnet := range r.Subnets() {
		result = append(result, r.routeTableAssociationsForSubnet(subnet.ID)...)
	}
	return cleanList(result)
}

func (r *AwsResources) InternetGateways() []ResourceTuple {
	result := []ResourceTuple{}
	igws, err := r.ec2client.DescribeInternetGateways(nil)
	if err != nil {
		fmt.Println("Error describing Internet Gateways:", err)
		return result
	}

	for _, igw := range igws.InternetGateways {
		url := INTERNET_GATEWAY_URL + *igw.InternetGatewayId
		result = append(result, ResourceTuple{
			ID:            *igw.InternetGatewayId,
			AwsConsoleURL: url,
		})
	}

	return cleanList(result)
}

func (r *AwsResources) HostedZones() []HostedZoneTuple {
	result := []HostedZoneTuple{}
	zones, err := r.route53client.ListHostedZones(nil)
	if err != nil {
		fmt.Println("Error listing hosted zones:", err)
		return result
	}

	for _, zone := range zones.HostedZones {
		id := strings.TrimSuffix(*zone.Name, ".")
		if contains(IGNORE_ZONE_NAMES, id) {
			continue
		}
		result = append(result, HostedZoneTuple{
			ID:           id,
			HostedZoneID: *zone.Id,
		})
	}

	return cleanList(result)
}

func (r *AwsResources) Rds() []ResourceTuple {
	result := []ResourceTuple{}
	marker := ""

	for {
		output, err := r.rdsclient.DescribeDBInstances(&rds.DescribeDBInstancesInput{
			Marker: &marker,
		})
		if err != nil {
			fmt.Println("Error describing DB instances:", err)
			return result
		}

		for _, db := range output.DBInstances {
			id := *db.DBInstanceIdentifier
			result = append(result, ResourceTuple{
				ID:            id,
				AwsConsoleURL: RDS_HOME + "#database:id=" + id + ";is-cluster=false",
			})
		}

		if output.Marker == nil {
			break
		}
		marker = *output.Marker
	}

	return result
}

func (r *AwsResources) RdsCluster() []ResourceTuple {
	result := []ResourceTuple{}
	marker := ""

	for {
		output, err := r.rdsclient.DescribeDBClusters(&rds.DescribeDBClustersInput{
			Marker: &marker,
		})
		if err != nil {
			fmt.Println("Error describing DB clusters:", err)
			return result
		}

		for _, db := range output.DBClusters {
			id := *db.DBClusterIdentifier
			result = append(result, ResourceTuple{
				ID:            id,
				AwsConsoleURL: RDS_HOME + "#database:id=" + id + ";is-cluster=true",
			})
		}

		if output.Marker == nil {
			break
		}
		marker = *output.Marker
	}

	return result
}

func (r *AwsResources) routeTablesForSubnet(subnetID string) []ResourceTuple {
	result := []ResourceTuple{}
	associations := r.routeTableAssociationObjects(subnetID)
	for _, rt := range associations {
		url := ROUTE_TABLE_URL + *rt.RouteTableId
		result = append(result, ResourceTuple{
			ID:            *rt.RouteTableId,
			AwsConsoleURL: url,
		})
	}
	return result
}

func (r *AwsResources) routeTableAssociationsForSubnet(subnetID string) []ResourceTuple {
	result := []ResourceTuple{}
	associations := r.routeTableAssociationObjects(subnetID)
	for _, rta := range associations {
		result = append(result, ResourceTuple{
			ID: *rta.RouteTableAssociationId,
		})
	}
	return result
}

func (r *AwsResources) routeTableAssociationObjects(subnetID string) []*ec2.RouteTableAssociation {
	result := []*ec2.RouteTableAssociation{}
	output, err := r.ec2client.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("association.subnet-id"),
				Values: []*string{aws.String(subnetID)},
			},
		},
	})
	if err != nil {
		fmt.Println("Error describing route tables:", err)
		return result
	}

	for _, rt := range output.RouteTables {
		result = append(result, rt.Associations...)
	}

	return result
}

func (r *AwsResources) subnetIds(vpcID string) []ResourceTuple {
	result := []ResourceTuple{}
	output, err := r.ec2client.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcID)},
			},
		},
	})
	if err != nil {
		fmt.Println("Error describing subnets:", err)
		return result
	}

	for _, sn := range output.Subnets {
		url := SUBNET_URL + *sn.SubnetId
		result = append(result, ResourceTuple{
			ID:            *sn.SubnetId,
			AwsConsoleURL: url,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

func (r *AwsResources) natGatewayIdsForVpc(vpcID string) []ResourceTuple {
	result := []ResourceTuple{}
	output, err := r.ec2client.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
		Filter: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcID)},
			},
		},
	})
	if err != nil {
		fmt.Println("Error describing NAT gateways:", err)
		return result
	}

	for _, ngw := range output.NatGateways {
		url := NAT_GATEWAY_URL + *ngw.NatGatewayId
		result = append(result, ResourceTuple{
			ID:            *ngw.NatGatewayId,
			AwsConsoleURL: url,
		})
	}

	return result
}

type ResourceTuple struct {
	ID            string
	AwsConsoleURL string
}

type HostedZoneTuple struct {
	ID           string
	HostedZoneID string
}

func cleanList[T any](list []T) []T {
	// Placeholder for the actual clean list logic
	return list
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getTagValue(tags []*ec2.Tag, key string) string {
	for _, tag := range tags {
		if *tag.Key == key {
			return *tag.Value
		}
	}
	return ""
}

func main() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
	}))

	resources := NewAwsResources(sess)

	fmt.Println("Kops Clusters:", resources.KopsClusters())
	fmt.Println("VPCs:", resources.Vpcs())
	fmt.Println("NAT Gateways:", resources.NatGateways())
	fmt.Println("Subnets:", resources.Subnets())
	fmt.Println("Route Tables:", resources.RouteTables())
	fmt.Println("Route Table Associations:", resources.RouteTableAssociations())
	fmt.Println("Internet Gateways:", resources.InternetGateways())
	fmt.Println("Hosted Zones:", resources.HostedZones())
	fmt.Println("RDS Instances:", resources.Rds())
	fmt.Println("RDS Clusters:", resources.RdsCluster())
}
