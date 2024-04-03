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
	HostedServices []HostedService `json:"namespace_details"`
	Total          []TotalCount
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

type TotalCount struct {
	TotalNamespaces   int
	TotalApplications int
}

func HostedServicesPage(w http.ResponseWriter, bucket string, client *s3.Client) {
	t := template.Must(template.ParseFiles("lib/templates/hosted_services.html"))

	// import json data from s3
	byteValue, err := utils.ImportS3File(client, bucket, "hosted_services.json")
	if err != nil {
		fmt.Println(err)
	}

	var hostedServices HostedServices

	json.Unmarshal(byteValue, &hostedServices)

	countNS := make(map[string]int)
	countApp := make(map[string]int)

	for i := 0; i < len(hostedServices.HostedServices); i++ {
		countNS[hostedServices.HostedServices[i].Namespace]++
		countApp[hostedServices.HostedServices[i].Application]++
	}

	var totalCount []TotalCount

	totalCount = append(totalCount, TotalCount{TotalNamespaces: len(countNS), TotalApplications: len(countApp)})

	// print total namespaces and applications count from struct
	for i := 0; i < len(hostedServices.Total); i++ {
		fmt.Printf("Total Namespaces: %d\n", hostedServices.Total[i].TotalNamespaces)
		fmt.Printf("Total Applications: %d\n", hostedServices.Total[i].TotalApplications)
	}

	fmt.Println("Namespace Details:")
	for i := 0; i < len(hostedServices.HostedServices); i++ {
		fmt.Printf("Namespace: %s\n", hostedServices.HostedServices[i].Namespace)
		fmt.Printf("Application: %s\n", hostedServices.HostedServices[i].Application)
		fmt.Printf("BusinessUnit: %s\n", hostedServices.HostedServices[i].BusinessUnit)
		fmt.Printf("TeamName: %s\n", hostedServices.HostedServices[i].TeamName)
		fmt.Printf("SlackChannel: %s\n", hostedServices.HostedServices[i].SlackChannel)
		fmt.Printf("SourceCode: %s\n", hostedServices.HostedServices[i].SourceCode)
		fmt.Printf("DomainNames: %s\n", hostedServices.HostedServices[i].DomainNames)
		fmt.Println("------------------------------")
	}

	// render template
	if err := t.ExecuteTemplate(w, "hosted_services.html", hostedServices); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
