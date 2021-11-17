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
	migratedSkipFilename = flag.String("migratedSkipFilename", "MIGRATED_SKIP_APPLY_THIS_NAMESPACE", "String of the aws weight annotation to check")
	hoodawApiKey         = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint       = flag.String("hoodawEndpoint", "/migrated_services", "Endpoint to send the data to")
	hoodawHost           = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	org                  = flag.String("org", "ministryofjustice", "GitHub user or organisation.")
	repository           = flag.String("repository", "cloud-platform-environments", "Repository to check the PR of.")
	token                = flag.String("token", os.Getenv("GITHUB_OAUTH_TOKEN"), "Personal access token for GitHub API.")

	endPoint = *hoodawHost + *hoodawEndpoint

	prPageCount       = 5
	nsMigratedBaseNum = 4

	// Based on live-1 user folders in the env repo as of 16 Nov and number of ns migrated to live
	live1NsMigrationPool = 358
)

func main() {
	flag.Parse()

	client, err := authenticate.GitHubClient(*token)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// query the list of files in all namespaces and add it to a slice
	nsPerDate, err := fetchNamespaceFolders(client)
	if err != nil {
		log.Fatalln(err.Error())

	}

	nsCountMigrated := perdayCount(nsPerDate)
	if err != nil {
		log.Fatalln(err.Error())

	}

	migratedMapSlice := buildMigratedSlice(nsCountMigrated)

	// for i, map := range migratedMapSlice {
	fmt.Println(migratedMapSlice)
	// }

	_, err = BuildJsonMap(migratedMapSlice)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// // Post json to hoowdaw api
	// err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	// if err != nil {
	// 	log.Fatalln(err.Error())
	// }

}

func fetchNamespaceFolders(client *github.Client) ([]map[string]string, error) {

	ns := make([]map[string]string, 0)
	// There is an assumption that the migration PRs are in the last 500 PRs which are in last 5 pages
	// Increase the page if you cannot get the full list of PRs related to migration

	for page := 1; page <= prPageCount; page++ {
		fmt.Println("Page:", page)
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

					if strings.Contains(*files.Filename, *migratedSkipFilename) && (*files.Status == "removed") {
						// namespaces filepaths are assumed to come in
						// the format: namespaces/live-1.cloud-platform.service.justice.gov.uk/<namespaceName>
						//s := strings.Split(*files.Filename, "/")

						m := make(map[string]string)
						//m["namespace"] = s[2]
						//mergedTime := pull.MergedAt.Format("2006-01-02")
						m["prmerged_date"] = pull.MergedAt.Format("2006-01-02")
						ns = append(ns, m)
					}
				}

			}
		}

	}

	return ns, nil
}

func perdayCount(nsPerDate []map[string]string) map[string]int {

	nsCountPerDate := make(map[string]int)

	for _, item := range nsPerDate {
		// check if the item/element exist in the duplicate_frequency map

		date := item["prmerged_date"]
		_, exist := nsCountPerDate[date]

		if exist {
			nsCountPerDate[date] += 1 // increase counter by 1 if already in the map
		} else {
			nsCountPerDate[date] = 1 // else start counting from 1
		}
	}

	//sort.Strings(nsCountPerDate)
	return nsCountPerDate

}

func buildMigratedSlice(nsCountPerDate map[string]int) []map[string]string {

	nsCountMigratedSlice := make([]map[string]string, 0)

	i, tillCount := 0, 0

	for date, nsCount := range nsCountPerDate {
		nsCountMigratedMap := make(map[string]string)

		if i == 0 {
			tillCount = nsMigratedBaseNum + nsCountPerDate[date]
		} else {
			tillCount += nsCount
		}
		i++

		fmt.Println("Date:", date, "Count:", nsCount)
		nsCountMigratedMap["date"] = date
		nsCountMigratedMap["todayCount"] = strconv.Itoa(nsCount)
		nsCountMigratedMap["tillCount"] = strconv.Itoa(tillCount)
		nsCountMigratedMap["percentage"] = fmt.Sprintf("%f", (float64(tillCount)/float64(live1NsMigrationPool))*100)
		nsCountMigratedSlice = append(nsCountMigratedSlice, nsCountMigratedMap)
	}

	return nsCountMigratedSlice
}

// BuildJsonMap takes a slice of maps and return a json encoded map
func BuildJsonMap(nsCountMigrated []map[string]string) ([]byte, error) {
	// To handle generics in the data type, we need to create a new map,
	// add the first key string:string and then the second key/value string:map[string]string.
	// As per the requirements of the HOODAW API.
	jsonMap := resourceMap{
		"updated_at":        time.Now().Format("2006-01-2 15:4:5 UTC"),
		"migrated_services": nsCountMigrated,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
