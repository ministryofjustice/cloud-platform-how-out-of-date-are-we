package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type HostedServices struct {
	HostedServices     []HostedService `json:"namespace_details"`
	TotalNamespaces    int
	UniqueApplications int
	LastUpdated        string
}

type HostedService struct {
	Namespace    string   `json:"Name"`
	Application  string   `json:"Application"`
	BusinessUnit string   `json:"BusinessUnit"`
	TeamName     string   `json:"TeamName"`
	SlackChannel string   `json:"TeamSlackChannel"`
	SourceCode   string   `json:"GithubURL"`
	DomainNames  []string `json:"DomainNames"`
}

func HostedServicesPage(w http.ResponseWriter, bucket string, client *s3.Client) {
	t := template.Must(template.ParseFiles("lib/templates/hosted_services.html"))

	// import json data from s3
	byteValue, filestamp, err := utils.ImportS3File(client, bucket, "hosted_services.json")
	if err != nil {
		fmt.Println(err)
	}

	var hostedServices HostedServices
	json.Unmarshal(byteValue, &hostedServices)

	hostedServices.LastUpdated = filestamp

	countNS := make(map[string]int)
	countApp := make(map[string]int)

	for i := 0; i < len(hostedServices.HostedServices); i++ {
		countNS[hostedServices.HostedServices[i].Namespace]++
		countApp[hostedServices.HostedServices[i].Application]++

		hostedServices.TotalNamespaces = len(countNS)
		hostedServices.UniqueApplications = len(countApp)
	}

	// render template
	if err := t.ExecuteTemplate(w, "hosted_services.html", hostedServices); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
