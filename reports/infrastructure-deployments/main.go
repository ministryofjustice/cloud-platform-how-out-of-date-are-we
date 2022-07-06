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
	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/migrated_services", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	org            = flag.String("org", "ministryofjustice", "GitHub user or organisation.")
	repository     = flag.String("repository", "cloud-platform-infrastructure", "Repository to check the PR of.")
	token          = flag.String("token", os.Getenv("GITHUB_OAUTH_TOKEN"), "Personal access token for GitHub API.")

	endPoint    = *hoodawHost + *hoodawEndpoint
	infraPath   = "terraform/aws-accounts/cloud-platform-aws"
	githubv4Url = "https://api.github.com/graphql"

	// Number of months to generate report
	numMonths = 3
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
	infraDeployed, infraFailed int
}

func main() {
	flag.Parse()

	// Set nthMonth to count 0 which is the current month.
	// This report generated data for past 12 months based on current month
	nthMonth := 0

	infraMonthMap := make(map[string]infraPRs, 0)
	for nthMonth < numMonths {

		date := getFirstLastDayofMonth(nthMonth)

		nodes, err := getPrsPerMonth(date, prCount)
		if err != nil {
			log.Fatalln(err.Error())
		}

		// query PRs that have changes under the infraPath
		infraPRs, err := getInfraPrsCount(nodes)
		if err != nil {
			log.Fatalln(err.Error())

		}
		infraMonthMap[date.monthIndex] = *infraPRs
		fmt.Println("Deployed", infraPRs.infraDeployed, "Failed:", infraPRs.infraFailed, "Date:", date.monthIndex)
		nthMonth++
	}

	jsonToPost, err := BuildJsonMap(infraMonthMap)
	if err != nil {
		log.Fatalln(err.Error())
	}

	str := string(jsonToPost)

	//Post json to hoowdaw api
	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}

}

func getFirstLastDayofMonth(nthMonth int) date {
	var d date
	Time := time.Now()
	t1 := Time.AddDate(0, -nthMonth, 0)
	year, month, _ := t1.Date()
	d.monthIndex = string(year) + "/" + string(month)

	d.first = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	d.last = time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)

	return d
}

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
				if strings.Contains(title, "revert") {
					infra.infraFailed++
				} else {
					infra.infraDeployed++
				}
			}
		}

	}

	return infra, nil
}

func BuildJsonMap(infraPRs map[string]infraPRs) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at":        time.Now().Format("2006-01-2 15:4:5 UTC"),
		"infra_deployments": infraPRs,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
