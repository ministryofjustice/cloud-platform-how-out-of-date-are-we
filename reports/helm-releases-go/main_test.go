package main

import (
	"reflect"
	"testing"
)

func Test_getNamespaces(t *testing.T) {
	type args struct {
		helmListJson string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "helm releases in namespaces",
			args: args{
				helmListJson: "[{\"namespace\":\"ns1\"},{\"namespace\":\"ns2\"}]",
			},
			want:    []string{"ns1", "ns2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNamespaces(tt.args.helmListJson)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_joinAllHelmReleases(t *testing.T) {
	type args struct {
		helmReleaseLive   []helmRelease
		helmReleaseManger []helmRelease
		helmReleaseLive_1 []helmRelease
	}
	tests := []struct {
		name string
		args args
		want []resourceMap
	}{
		{
			name: "helm releases in namespaces",
			args: args{
				helmReleaseLive: []helmRelease{
					{
						Name:             "live",
						Namespace:        "live",
						InstalledVersion: "10",
						LatestVersion:    "11",
						Chart:            "live",
					},
				},
				helmReleaseManger: []helmRelease{
					{
						Name:             "manager",
						Namespace:        "manager",
						InstalledVersion: "10",
						LatestVersion:    "11",
						Chart:            "manager",
					},
				},
				helmReleaseLive_1: []helmRelease{
					{
						Name:             "live-1",
						Namespace:        "live-1",
						InstalledVersion: "10",
						LatestVersion:    "11",
						Chart:            "live-1",
					},
				},
			},
			want: []resourceMap{
				{
					"name": "live",
					"apps": []helmRelease{
						{
							Name:             "live",
							Namespace:        "live",
							InstalledVersion: "10",
							LatestVersion:    "11",
							Chart:            "live",
						},
					},
				},
				{
					"name": "manager",
					"apps": []helmRelease{
						{
							Name:             "manager",
							Namespace:        "manager",
							InstalledVersion: "10",
							LatestVersion:    "11",
							Chart:            "manager",
						},
					},
				},
				{
					"name": "live-1",
					"apps": []helmRelease{
						{
							Name:             "live-1",
							Namespace:        "live-1",
							InstalledVersion: "10",
							LatestVersion:    "11",
							Chart:            "live-1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinAllHelmReleases(tt.args.helmReleaseLive, tt.args.helmReleaseManger, tt.args.helmReleaseLive_1); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("joinAllHelmReleases() = %v, want %v", got, tt.want)
			}
		})
	}
}
