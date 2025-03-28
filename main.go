package main

import (
	"fmt"
	"log"
	"net/http"

	lib "github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/lib"
	utils "github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

var (
	bucket        = "cloud-platform-hoodaw-reports"
	errorNsBucket = "cloud-platform-concourse-environments-live-reports"
)

func main() {
	client, err := utils.S3Client("eu-west-2")
	if err != nil {
		fmt.Println(err)
	}

	for _, b := range []string{bucket, errorNsBucket} {
		exists, err := utils.CheckBucketExists(client, b)
		if err != nil {
			fmt.Println(err)
		}

		if !exists {
			fmt.Println("Bucket does not exist")
		}
	}

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("lib/static"))))

	http.HandleFunc("/hosted_services", func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		wantJson := accept == "application/json"
		lib.HostedServicesPage(w, bucket, wantJson, client)
	})

	http.HandleFunc("/helm_whatup", func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		wantJson := accept == "application/json"
		lib.HelmReleasesPage(w, bucket, wantJson, client)
	})

	http.HandleFunc("/costs_by_namespace", func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		wantJson := accept == "application/json"
		lib.NamespaceCostsPage(w, bucket, wantJson, client)
	})

	http.HandleFunc("/erroring_namespaces", func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		wantJson := accept == "application/json"
		lib.ErroredNamespacesPage(w, errorNsBucket, wantJson, client)
	})

	http.HandleFunc("GET /namespace/{namespace}", func(w http.ResponseWriter, r *http.Request) {
		namespace := r.PathValue("namespace")
		accept := r.Header.Get("Accept")
		wantJson := accept == "application/json"
		lib.NamespaceUsagePage(w, bucket, namespace, wantJson, client)
	})

	http.HandleFunc("GET /live_one_domains", func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		wantJson := accept == "application/json"
		lib.LiveOneDomainsPage(w, bucket, wantJson, client)
	})

	fmt.Println("Listening on port :8080 ...")
	serverErr := http.ListenAndServe(":8080", nil)
	if serverErr != nil {
		log.Fatal("Error starting server: ", serverErr)
	}
}
