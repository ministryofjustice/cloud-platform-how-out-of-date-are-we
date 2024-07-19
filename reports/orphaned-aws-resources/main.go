package main

var (
	git_source = "https://github.com/cloud-platform-environments"
)

var List = map[string][]string{
	"aws-sdk-rds":         {},
	"aws-sdk-ec2":         {},
	"aws-sdk-s3":          {},
	"aws-sdk-autoscaling": {"~> 1.44"},
	"aws-sdk-route53":     {"~> 1.40"},
	"slack-ruby-client":   {},
	"slack-notifier":      {},
	"json":                {">= 2.3.1"},
}
