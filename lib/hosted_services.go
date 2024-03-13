package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

func HostedServices(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("lib/templates/hosted_services.html"))

	// import json data from s3
	data, err := utils.ImportS3File("hosted-services", "hosted_services.json")
	if err != nil {
		fmt.Println(err)
	}

	// parse json data
	var services []utils.HostedService
	err = json.Unmarshal(data, &services)
	if err != nil {
		fmt.Println(err)
	}

	// render template
	if err := t.ExecuteTemplate(w, "rides.html", services); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
