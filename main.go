package main

import (
	"fmt"
	"net/http"

	lib "github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/lib"
	utils "github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/utils"
)

var (
	bucket = "cloud-platform-hoodaw-reports"
)

func main() {
	client, err := utils.S3Client()
	if err != nil {
		fmt.Println(err)
	}
	exists, err := utils.CheckBucketExists(client, bucket)
	if err != nil {
		fmt.Println(err)
	}

	if !exists {
		fmt.Println("Bucket does not exist")
	}

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("lib/static"))))
	http.HandleFunc("/hosted-services", func(w http.ResponseWriter, r *http.Request) {
		lib.HostedServices(w, r, bucket, client)
	})
	fmt.Println("Listening")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
