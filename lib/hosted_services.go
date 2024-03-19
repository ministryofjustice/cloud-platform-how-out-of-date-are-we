package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type HostedService struct {
	Namespace    string `json:"namespace"`
	Application  string `json:"application"`
	BusinessUnit string `json:"business_unit"`
	TeamName     string `json:"team_name"`
	SlackChannel string `json:"slack_channel"`
	SourcesCode  string `json:"source_code"`
	DomainName   string `json:"domain_name"`

	TotalNS,
	TotalApps int
}

func HostedServices(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("lib/templates/hosted_services.html"))

	// import json data from s3
	data, err := utils.ImportS3File("hosted-services", "hosted_services.json")
	if err != nil {
		fmt.Println(err)
	}

	// parse json data
	var services []HostedService
	err = json.Unmarshal(data, &services)
	if err != nil {
		fmt.Println(err)
	}

	// count the number of namespaces and applications in the json data and add to the struct
	for i, _ := range services {
		services[i].TotalNS = len(services[i].Namespace)
		services[i].TotalApps = len(services[i].Application)
	}

	// render template
	if err := t.ExecuteTemplate(w, "rides.html", services); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
