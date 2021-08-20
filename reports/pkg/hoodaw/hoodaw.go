package hoodaw

import (
	"bytes"
	"net/http"
)

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
