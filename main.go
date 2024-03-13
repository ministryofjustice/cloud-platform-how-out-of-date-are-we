package main

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/cloud-platform-how-out-of-date-are-we/lib"
)

func main() {
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("lib/static"))))
	http.HandleFunc("/", lib.HostedServices)
	fmt.Println("Listening")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
