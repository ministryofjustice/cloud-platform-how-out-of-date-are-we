package utils

// this needs ajusting to match the json data structure from the s3 bucket file
type HostedService struct {
	Namespacce   string `json:"namespace"`
	Application  string `json:"application"`
	BusinessUnit string `json:"business_unit"`
	TeamName     string `json:"team_name"`
	SlackChannel string `json:"slack_channel"`
	SourcesCode  string `json:"source_code"`
	Domainname   string `json:"domain_name"`
}
