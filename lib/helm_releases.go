package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type Cluster struct {
	HelmReleases []HelmRelease `json:"apps"`
	ClusterName  string        `json:"name"`
}

type HelmReleases struct {
	Clusters    []Cluster `json:"clusters"`
	LastUpdated string
}

type HelmRelease struct {
	Name             string `json:"name"`
	Chart            string `json:"chart"`
	Namespace        string `json:"namespace"`
	InstalledVersion string `json:"installed_version"`
	LatestVersion    string `json:"latest_version"`
	State            string
}

func HelmReleasesPage(w http.ResponseWriter, bucket string, client *s3.Client) {
	t := template.Must(template.ParseFiles("lib/templates/helm_releases.html"))

	byteValue, filestamp, err := utils.ImportS3File(client, bucket, "helm_releases.json")
	if err != nil {
		fmt.Println(err)
	}

	var helmReleases HelmReleases
	json.Unmarshal(byteValue, &helmReleases)

	helmReleases.LastUpdated = filestamp

	for i, c := range helmReleases.Clusters {
		for j, h := range c.HelmReleases {
			helmReleases.Clusters[i].HelmReleases[j].State = utils.CompareVersions(h.InstalledVersion, h.LatestVersion)
		}
	}

	if err := t.ExecuteTemplate(w, "helm_releases.html", helmReleases); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
