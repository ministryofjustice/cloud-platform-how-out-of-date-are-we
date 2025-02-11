package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type Breakdown struct{}

type NamespaceCost struct {
	Breakdown map[string]float32 `json:"breakdown"`
	Total     float32
}

type Costs struct {
	Namespaces  map[string]NamespaceCost `json:"namespace"`
	LastUpdated string
	Total       float32
}

func NamespaceCostsPage(w http.ResponseWriter, bucket string, wantJson bool, client *s3.Client) {
	t := template.Must(template.ParseFiles("lib/templates/namespace_costs.html"))

	byteValue, filestamp, err := utils.ImportS3File(client, bucket, "namespace_costs.json")
	if err != nil {
		fmt.Println(err)
	}

	if wantJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write(byteValue)
		return
	}

	var namespaceCosts Costs
	json.Unmarshal(byteValue, &namespaceCosts)

	namespaceCosts.LastUpdated = filestamp

	for _, ns := range namespaceCosts.Namespaces {
		namespaceCosts.Total += ns.Total
	}

	if err := t.ExecuteTemplate(w, "namespace_costs.html", namespaceCosts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
