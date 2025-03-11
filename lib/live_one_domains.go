package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type Domains struct {
	Data []struct {
		Namespace         string `json:"namespace"`
		IngressName       string `json:"ingress"`
		URL               string `json:"hostname"`
		CreationTimestamp string `json:"CreatedAt"`
	} `json:"live_one_domains"`
	LastUpdated string `json:"last_updated"`
	Total       int
}

func LiveOneDomainsPage(w http.ResponseWriter, bucket string, wantJson bool, client *s3.Client) {
	t := template.Must(template.ParseFiles("lib/templates/live_one_domains.html"))

	byteValue, filestamp, err := utils.ImportS3File(client, bucket, "live_one_domains.json")
	if err != nil {
		fmt.Println(err)
	}

	if wantJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write(byteValue)
		return
	}

	var domains Domains
	json.Unmarshal(byteValue, &domains)

	domains.LastUpdated = filestamp
	domains.Total = len(domains.Data)

	if err := t.ExecuteTemplate(w, "live_one_domains.html", domains); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
