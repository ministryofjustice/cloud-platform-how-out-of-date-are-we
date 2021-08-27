package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/reports/pkg/hoodaw"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes namespaces and hostnames that don't contain the annotation contained within
// the variable 'annotation'.
type ResourceMap map[string]interface{}

var (
	currentTime = time.Now()

	hoodawApiKey   = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawEndpoint = flag.String("hoodawEndpoint", "/documentation", "Endpoint to send the data to")
	hoodawHost     = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
	runBook        = flag.String("runBook", "runbooks.cloud-platform.service.justice.gov.uk", "Full URL of the runbook site.")
	userGuide      = flag.String("userGuide", "user-guide.cloud-platform.service.justice.gov.uk", "Full URL of the userguide.")

	endPoint = *hoodawHost + *hoodawEndpoint
)

func main() {
	flag.Parse()

	jsonToPost, err := collect()
	if err != nil {
		log.Fatalln("Failed to collect expired links:", err)
	}

	err = hoodaw.PostToApi(jsonToPost, hoodawApiKey, &endPoint)
	if err != nil {
		log.Fatalln("Failed to post to hoodaw:", err)
	}
}

func collect() ([]byte, error) {
	c := colly.NewCollector(
		colly.AllowedDomains(*userGuide, *runBook),
		colly.Async(true),
		colly.UserAgent("How-out-of-date-are-we/documentation"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "justice",
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})

	// Find and visit all links on the parent page.
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	// s contains a slice of maps, each map will be iterated over when placed in a dashboard.
	s := make([]map[string]string, 0)

	// Look for div value "data-last-reviewed-on" which contains an int value
	c.OnHTML("div[data-last-reviewed-on]", func(e *colly.HTMLElement) {
		lastReviewed, _ := e.DOM.Attr("data-last-reviewed-on")
		page := e.Request.URL.String()
		title := strings.Split(page, "/")

		// Add the page url as a key and the date of last review as a value.
		if lastReviewed < currentTime.Format("2006-01-02") {
			expired := map[string]string{
				"url":   page,
				"title": title[len(title)-1],
				"site":  title[2],
			}
			s = append(s, expired)
		}
	})

	c.Visit("https://" + *userGuide)
	c.Visit("https://" + *runBook)

	c.Wait()

	jsonMap := ResourceMap{
		"updated_at": time.Now().Format("2006-01-2 15:4:5 UTC"),
		// Adding z to the string name ensures the first key listed will be "updated_at", as
		// required by the HOODAW API.
		"pages": s,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
