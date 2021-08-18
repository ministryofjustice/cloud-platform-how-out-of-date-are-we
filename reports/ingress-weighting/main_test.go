package main

import (
	"testing"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func TestFromS3Bucket(t *testing.T) {
	type args struct {
		bucket     string
		configFile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "environment vars set",
			args: args{
				bucket:     "cloud-platform-concourse-kubeconfig",
				configFile: "live-1-only",
			},
			wantErr: false,
		},
		{
			name: "wrong bucket and config file",
			args: args{
				bucket:     "doesn't-exist-probably",
				configFile: "nofile",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromS3Bucket(tt.args.bucket, tt.args.configFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromS3Bucket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_postToApi(t *testing.T) {
	var (
		hoodawApiKey = "soopersekrit"
		endPoint     = "http://localhost:4567/test_endpoint"
	)
	type args struct {
		jsonToPost   []byte
		hoodawApiKey *string
		endPoint     *string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Post random json to localhost",
			args: args{
				jsonToPost:   []byte{'A'},
				hoodawApiKey: &hoodawApiKey,
				endPoint:     &endPoint,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := postToApi(tt.args.jsonToPost, tt.args.hoodawApiKey, tt.args.endPoint); (err != nil) != tt.wantErr {
				t.Errorf("postToApi() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
