package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes namespaces and hostnames that don't contain the annotation contained within
// the variable 'annotation'.
type resourceMap map[string]interface{}

var (
	hoodawBucket   = flag.String("howdaw-bucket", os.Getenv("HOODAW_BUCKET"), "AWS S3 bucket for hoodaw json reports")
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/infrastructure_deployments", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	org            = flag.String("org", "ministryofjustice", "GitHub user or organisation.")
	repository     = flag.String("repository", "cloud-platform-infrastructure", "Repository to check the PR of.")
	token          = flag.String("token", os.Getenv("GITHUB_OAUTH_TOKEN"), "Personal access token for GitHub API.")
	endPoint       = *hoodawHost + *hoodawEndpoint
)

const (
	infraPath = "terraform/aws-accounts/cloud-platform-aws"

	// Number of months to generate report
	numMonths = 12
	// number of PRs to fetch per month. set to 100 with assumption
	// that there are no more than 100 prs in the infrastructure repo
	prCount = 100
)

type nodes struct {
	PullRequest struct {
		Title githubv4.String
		Url   githubv4.String
	} `graphql:"... on PullRequest"`
}

type date struct {
	first      time.Time
	last       time.Time
	monthIndex string
}

type infraPRs struct {
	deployed, failed int
}

func main() {
	flag.Parse()

	infraReport := make([]map[string]string, 0)

	// Start from m = 0 which is the current month.
	// This report generated data for past 12 months based on current month
	for m := 0; m < numMonths; m++ {
		infraPRMap := make(map[string]string)

		date := getFirstLastDayofMonth(m)
		nodes, err := getPrsPerMonth(date, prCount)
		if err != nil {
			log.Fatalln(err.Error())
		}
		// query PRs that have changes under the infraPath
		infraPRs, err := getInfraPrsCount(nodes)
		if err != nil {
			log.Fatalln(err.Error())
		}
		infraPRMap["date"] = date.monthIndex
		infraPRMap["deployed"] = strconv.Itoa(infraPRs.deployed)
		infraPRMap["failed"] = strconv.Itoa(infraPRs.failed)
		infraReport = append(infraReport, infraPRMap)
	}
	jsonToPost, err := BuildJsonMap(infraReport)
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

	utils.ExportToS3(client, *hoodawBucket, "infratructure_deployments.json", jsonToPost)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Post json to hoowdaw api
	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// getFirstLastDayofMonth takes the month as input and return first day, last day and the index formatted
// for using in report
func getFirstLastDayofMonth(nthMonth int) date {
	var d date
	Time := time.Now()
	t1 := Time.AddDate(0, -nthMonth, 0)
	year, month, _ := t1.Date()
	d.monthIndex = strconv.Itoa(year) + "/" + month.String()
	d.first = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	d.last = time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)

	return d
}

// getPrsPerMonth takes date and number of PRs count as input, search the github using Graphql api for
//
//	list of PRs (title,url) between the first and last day provided
func getPrsPerMonth(date date, count int) ([]nodes, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewEnterpriseClient("https://api.github.com/graphql", httpClient)

	var query struct {
		Search struct {
			Nodes []nodes
		} `graphql:"search(first: $count, query: $searchQuery, type: ISSUE)"`
	}

	variables := map[string]interface{}{
		"searchQuery": githubv4.String(fmt.Sprintf(`repo:ministryofjustice/cloud-platform-infrastructure is:pr is:closed merged:%s..%s`, date.first.Format("2006-01-02"), date.last.Format("2006-01-02"))),
		"count":       githubv4.Int(count),
	}

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, err
	}

	return query.Search.Nodes, nil
}

// getInfraPrsCount get the list of github PullRequests and calls github REST API and get the list of changed files
// of each PR. Check if the changed file is under the terraform path and increment the deployed counter.
// It also checks if the PR title has word "revert" then increment the failed counter.
func getInfraPrsCount(nodes []nodes) (*infraPRs, error) {
	infra := new(infraPRs)

	// Authenticate to github using auth token
	client, err := authenticate.GitHubClient(*token)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, pr := range nodes {
		url := string(pr.PullRequest.Url)
		prNumber, err := strconv.Atoi(url[strings.LastIndex(url, "/")+1:])
		if err != nil {
			return nil, nil
		}
		title := string(pr.PullRequest.Title)
		commitFiles, _, _ := client.PullRequests.ListFiles(context.Background(), *org, *repository, prNumber, nil)
		for _, files := range commitFiles {
			// consider only the changes under infraPath
			if strings.Contains(*files.Filename, infraPath) {
				// This is an assumption if PR titles have revert in it, it is
				// a revert of a previous failed deployment
				// Should the Success deployments count be decremented?????
				if strings.Contains(strings.ToLower(title), "revert") {
					infra.failed++
				} else {
					infra.deployed++
				}
				break
			}
		}

	}

	return infra, nil
}

// BuildJsonMap takes a map with date key and infraPRs struct as value, and return a json encoded map
func BuildJsonMap(infraPRs []map[string]string) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at":  time.Now().Format("2006-01-2 15:4:5 UTC"),
		"deployments": infraPRs,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
