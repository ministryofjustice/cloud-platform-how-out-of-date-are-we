package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type HostedService struct {
	Namespace    string `json:"Name"`
	Application  string `json:"Application"`
	BusinessUnit string `json:"BusinessUnit"`
	TeamName     string `json:"TeamName"`
	SlackChannel string `json:"TeamSlackChannel"`
	SourceCode   string `json:"GithubURL"`
	DomainNames  string `json:"DomainNames"`

	TotalNS,
	TotalApps int
}

var (
	totalNamespace,
	totalApplication int
)

func HostedServices(w http.ResponseWriter, r *http.Request, bucket string, client *s3.Client) {
	t := template.Must(template.ParseFiles("lib/templates/hosted_services.html"))

	// import json data from s3
	var services []HostedService
	data, err := utils.ImportS3File(client, bucket, "hosted_services.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(data, &services)
	if err != nil {
		fmt.Println(err)
	}

	// count total number of namespaces and applications in struct
	for _, service := range services {
		totalNamespace = len(service.Namespace)
		totalApplication = len(service.Application)
	}
	services = append(services, HostedService{TotalNS: totalNamespace, TotalApps: totalApplication})

	// render template
	if err := t.ExecuteTemplate(w, "hosted_services.html", services); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
