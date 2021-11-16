package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/ministryofjustice/cloud-platform-environments/pkg/authenticate"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes namespaces and hostnames that don't contain the annotation contained within
// the variable 'annotation'.
// type resourceMap map[string]interface{}

var (
	migratedSkipFilename = flag.String("migratedSkipFilename", "MIGRATED_SKIP_APPLY_THIS_NAMESPACE", "String of the aws weight annotation to check")
	//	hoodawApiKey         = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	//	hoodawEndpoint = flag.String("hoodawEndpoint", "/migrated_services", "Endpoint to send the data to")
	//	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	org        = flag.String("org", "ministryofjustice", "GitHub user or organisation.")
	repository = flag.String("repository", "cloud-platform-environments", "Repository to check the PR of.")
	token      = flag.String("token", os.Getenv("GITHUB_OAUTH_TOKEN"), "Personal access token for GitHub API.")

//	endPoint = *hoodawHost + *hoodawEndpoint
)

func main() {
	flag.Parse()

	client, err := authenticate.GitHubClient(*token)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// query the list of files in all namespaces and add it to a slice
	namespaceNames, err := fetchNamespaceFolders(client)
	if err != nil {
		log.Fatalln(err.Error())

	}

	for _, ns := range namespaceNames {
		fmt.Println(ns["prmergedtime"], ",", ns["namespace"])
	}

}

func fetchNamespaceFolders(client *github.Client) ([]map[string]string, error) {

	ns := make([]map[string]string, 0)
	// There is an assumption that the migration PRs are in the last 500 PRs which are in last 30 pages
	// Increase the page if you cannot get the full list of PRs related to migration

	for page := 1; page <= 5; page++ {
		fmt.Println("page", page)
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

			if pull.MergedAt != nil {

				commitFiles, _, _ := client.PullRequests.ListFiles(context.Background(), *org, *repository, *pull.Number, nil)

				for _, files := range commitFiles {

					if strings.Contains(*files.Filename, *migratedSkipFilename) && (*files.Status == "removed") {
						// namespaces filepaths are assumed to come in
						// the format: namespaces/live-1.cloud-platform.service.justice.gov.uk/<namespaceName>
						s := strings.Split(*files.Filename, "/")

						m := make(map[string]string)
						m["namespace"] = s[2]
						//mergedTime := pull.MergedAt.Format("2006-01-02")
						m["prmergedtime"] = pull.MergedAt.String()
						ns = append(ns, m)
					}
				}

			}
		}

	}

	return ns, nil
}
