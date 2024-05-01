package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

type HelmReleases struct {
	HelmReleases []HelmRelease `json:"apps"`
	LastUpdated  string
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

	// import json data from s3
	byteValue, filestamp, err := utils.ImportS3File(client, bucket, "helm_releases.json")
	if err != nil {
		fmt.Println(err)
	}

	var helmReleases HelmReleases
	json.Unmarshal(byteValue, &helmReleases)

	helmReleases.LastUpdated = filestamp

	// compare installed and latest versions of helm releases traffic light system
	for i := 0; i < len(helmReleases.HelmReleases); i++ {
		helmReleases.HelmReleases[i].State = utils.CompareVersions(helmReleases.HelmReleases[i].InstalledVersion, helmReleases.HelmReleases[i].LatestVersion)
	}

	// render template
	if err := t.ExecuteTemplate(w, "helm_releases.html", helmReleases); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
