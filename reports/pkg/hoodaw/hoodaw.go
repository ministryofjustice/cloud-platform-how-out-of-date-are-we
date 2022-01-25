package hoodaw

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"
)

const host = "https://reports.cloud-platform.service.justice.gov.uk/"

// resourceMap is used to store both string:string and string:map[string]string key
// value pairs. The HOODAW API requires the first entry of map to contain a string:string,
// the rest of the map consists of a primary key (string) with a value containing a map (string:string)
// of Kubernetes namespaces and hostnames that don't contain the annotation contained within
// the variable 'annotation'.
type ResourceMap map[string]interface{}

// QueryApi takes the name of an endpoint and uses the http package to
// return a slice of bytes.
func QueryApi(endPoint string) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, host+endPoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "hoodaw-pkg")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// postToApi takes a slice of bytes as an argument and attempts to POST it to the REST API
// provided by HOODAW. The slice of bytes should contain json using the guidelines outlined
// by HOODAW i.e. the first entry in the key value pair should contain a string:string, which consists
// of a string and the time POSTed.
func PostToApi(jsonToPost []byte, hoodawApiKey, endPoint *string) error {
	req, err := http.NewRequest("POST", *endPoint, bytes.NewBuffer(jsonToPost))
	if err != nil {
		return err
	}

	req.Header.Add("X-API-KEY", *hoodawApiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
