package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gocolly/colly"
)

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes namespaces and hostnames that don't contain the annotation contained within
// the variable 'annotation'.
type ResourceMap map[string]interface{}

var (
	userGuide   = flag.String("userGuide", "user-guide.cloud-platform.service.justice.gov.uk", "Full URL of the userguide.")
	runBook     = flag.String("runBook", "runbooks.cloud-platform.service.justice.gov.uk", "Full URL of the runbook site.")
	currentTime = time.Now()
)

func main() {
	flag.Parse()

	expired, err := collect()
	if err != nil {
		log.Fatalln("Failed to collect expired links:", err)
	}

	fmt.Println(string(expired))
}

func collect() ([]byte, error) {
	// spider url looking for links to other pages
	// return a hash of pages: { pageUrl : needsReview? }
	c := colly.NewCollector(
		colly.AllowedDomains(*userGuide, *runBook),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "justice",
		Parallelism: 2,
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
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
		// Add the page url as a key and the date of last review as a value.
		if lastReviewed < currentTime.Format("2006-01-02") {
			expired := make(map[string]string)
			expired[page] = lastReviewed
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
		"zExpired": s,
	}

	jsonStr, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
