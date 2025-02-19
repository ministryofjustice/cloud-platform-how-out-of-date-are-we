package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type NamespaceError struct {
	Namespace string `json:"namespace"`
	Error     string `json:"error"`
}

type ErroringNamespaces struct {
	Namespaces  []NamespaceError `json:"namespaces"`
	LastUpdated string
}

func ErroredNamespacesPage(w http.ResponseWriter, bucket string, wantJson bool, client *s3.Client) {
	t := template.Must(template.ParseFiles("lib/templates/erroring_namespaces.html"))

	byteValue, filestamp, err := utils.ImportS3File(client, bucket, "apply-live/collated-errored-namespaces.json")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to load data from S3", http.StatusInternalServerError)
		return
	}

	if wantJson {
		w.Header().Set("Content-Type", "application/json")
		w.Write(byteValue)
		return
	}

	var erroringNamespaces []NamespaceError
	if err := json.Unmarshal(byteValue, &erroringNamespaces); err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to parse JSON data", http.StatusInternalServerError)
		return
	}

	data := ErroringNamespaces{
		Namespaces:  erroringNamespaces,
		LastUpdated: filestamp,
	}

	if err := t.ExecuteTemplate(w, "erroring_namespaces.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
