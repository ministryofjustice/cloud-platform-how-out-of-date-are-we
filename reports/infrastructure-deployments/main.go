package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
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

	endPoint  = *hoodawHost + *hoodawEndpoint
	infraPath = "terraform/aws-accounts/cloud-platform-aws"

	// Nunber of pages of PRs to look in github
	prPageCount = 1
)

func main() {
	flag.Parse()

	n := 0
	for n < 12 {
		Time := time.Now()
		t1 := Time.AddDate(0, -n, 0)
		y, m, _ := t1.Date()
		first, last := monthInterval(y, m)

		prsInMonth, err := fetchPrsPerMonth(first, last)
		if err != nil {
			log.Fatalln(err.Error())
		}

		n++

		fmt.Println(prsInMonth)

	}

	// Authenticate to github using auth token
	client, err := authenticate.GitHubClient(*token)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// query PRs that have changes under the infraPath
	infraChangedDates, err := fetchInfraChangedDates(client)
	if err != nil {
		log.Fatalln(err.Error())

	}

	// group the dates by month
	nsCountChanged := perMonthCount(infraChangedDates)
	if err != nil {
		log.Fatalln(err.Error())

	}

	for key, element := range nsCountChanged {
		fmt.Println("Month:", key, "=>", "deployments:", element)
	}

	// Build the migrated report slice
	migratedMapSlice := buildMigratedSlice(nsCountChanged)

	jsonToPost, err := BuildJsonMap(migratedMapSlice)
	if err != nil {
		log.Fatalln(err.Error())
	}

	str := string(jsonToPost)
	fmt.Print(str)

	Post json to hoowdaw api
	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln(err.Error())
	}

}

func fetchPrsPerMonth(first, last time.Time) ([]string, error) {
	query := fmt.Sprintf(`{ search(first: 100, query: "repo:ministryofjustice/cloud-platform-infrastructure is:pr is:closed merged:%s..%s", type: ISSUE )
			 { nodes { ... on PullRequest { url }}}}`, first.Format("2006-01-02"), last.Format("2006-01-02"))

	b, err := json.Marshal(struct {
		Query    string                 `json:"query"`
		Variable map[string]interface{} `json:"variables"`
	}{
		Query: query,
		Variable: map[string]interface{}{
			"login": "github",
		},
	})
	if err != nil {
		return nil, err
	}

	endpointURL, err := url.Parse("https://api.github.com/graphql")
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(b)
	resp, err := http.DefaultClient.Do(&http.Request{
		URL:    endpointURL,
		Method: "POST",
		Header: http.Header{
			"Content-Type":  {"application/json"},
			"Authorization": {"bearer " + *token},
		},
		Body: ioutil.NopCloser(buf),
	})
	if err != nil {
		return nil, err
	}
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type search struct {
		Nodes []map[string]string `json:"nodes"`
	}
	type data struct {
		Search search `json:"search"`
	}
	type response struct {
		Data data `json:"data"`
	}

	var respJson response

	json.Unmarshal([]byte(b), &respJson)
	nodes := respJson.Data.Search.Nodes

	prNumbers := make([]string, 0)

	for _, urlMap := range nodes {
		for _, url := range urlMap {
			prNumber := url[strings.LastIndex(url, "/")+1:]

			prNumbers = append(prNumbers, prNumber)

		}
	}
}

// fetchMigratedDates paginate through the PRs, filter down the ones that are merged
// Then, it list the changedFiles and check if migratedSkip file is deleted in that PR.
// This represents that the namespace is deleted from `live-1` in that PR.
// It then added the merged date to a slice of strings and return the same
func fetchInfraChangedDates(client *github.Client) ([]string, error) {

	infraChangedDates := make([]string, 0)
	// There is an assumption that the migration PRs are in the last 500 PRs which are in last 5 pages
	// Increase the page if you cannot get the full list of PRs related to migration

	for page := 1; page <= prPageCount; page++ {
		opts := &github.PullRequestListOptions{
			State:       "closed",
			Sort:        "updated",
			Direction:   "desc",
			ListOptions: github.ListOptions{Page: page, PerPage: 100},
		}
		ctx := context.Background()
		closedPulls, _, err := client.PullRequests.List(ctx, *org, *repository, opts)
		if err != nil {
			return nil, err
		}

		for _, pull := range closedPulls {
			// Check for MergedAt to filter only PRs that are merged and not closed
			if pull.MergedAt != nil {

				commitFiles, _, _ := client.PullRequests.ListFiles(context.Background(), *org, *repository, *pull.Number, nil)

				for _, files := range commitFiles {
					// Group the dates by month and add count per month
					if strings.Contains(*files.Filename, infraPath) {
						// TODO Check for "Revert" in the commit/PR and add a seperate list for failed deployment
						infraChangedDates = append(infraChangedDates, pull.MergedAt.Format("2006-01-02"))
						break
					}
				}

			}
		}

	}

	return infraChangedDates, nil
}

// perMonthCount get the slice of dates and group them by month and add the number of occurance of entry
// in to a map. The occurence of a month represents a deployment on that month
func perMonthCount(infraChangedDates []string) map[string]int {

	nsCountPerMonth := make(map[string]int)

	for _, date := range infraChangedDates {

		month := string([]rune(date[5:7]))
		year := string([]rune(date[0:4]))

		key := month + "/" + year

		// check if the item/element exist in the nsCountPerMonth map
		_, exist := nsCountPerMonth[key]

		if exist {
			nsCountPerMonth[key] += 1 // increase counter by 1 if already in the map
		} else {
			nsCountPerMonth[key] = 1 // else start counting from 1
		}
	}

	return nsCountPerMonth

}

// buildMigratedSlice gets a map of months, sort the months and build a slice of maps
// with details required for the report
func buildMigratedSlice(nsCountPerMonth map[string]int) []map[string]string {

	nsCountChangedSlice := make([]map[string]string, 0)

	tillCount := 0

	// sort the months before building the Slice so the months are in asc order
	keys := make([]string, 0, len(nsCountPerMonth))
	for k := range nsCountPerMonth {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, monthYear := range keys {
		nsCountChangedMap := make(map[string]string)
		tillCount += nsCountPerMonth[monthYear]

		nsCountChangedMap["month"] = monthYear
		nsCountChangedMap["MonthCount"] = strconv.Itoa(nsCountPerMonth[monthYear])
		nsCountChangedMap["tillCount"] = strconv.Itoa(tillCount)
		nsCountChangedSlice = append(nsCountChangedSlice, nsCountChangedMap)
	}

	return nsCountChangedSlice
}

// // BuildJsonMap takes a slice of maps and return a json encoded map
func BuildJsonMap(nsCountChanged []map[string]string) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at":        time.Now().Format("2006-01-2 15:4:5 UTC"),
		"infra_deployments": nsCountChanged,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}

func monthInterval(y int, m time.Month) (firstDay, lastDay time.Time) {
	firstDay = time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	lastDay = time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC)
	return firstDay, lastDay
}
