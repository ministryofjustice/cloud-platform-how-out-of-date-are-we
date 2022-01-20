package main

import (
	"reflect"
	"testing"

	"github.com/ministryofjustice/cloud-platform-environments/pkg/namespace"
	v1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetNamespaceDetails(t *testing.T) {
	type args struct {
		ns v1.Namespace
	}
	tests := []struct {
		name string
		args args
		want namespace.Namespace
	}{
		{
			name: "ns1",
			args: args{
				ns: v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ns1",
						Labels: map[string]string{
							"cloud-platform.justice.gov.uk/environment-name": "test",
							"cloud-platform.justice.gov.uk/is-production":    "false",
						},
						Annotations: map[string]string{
							"cloud-platform.justice.gov.uk/application":   "test-app",
							"cloud-platform.justice.gov.uk/business-unit": "test-bu",
							"cloud-platform.justice.gov.uk/owner:Digital": "Test Services: test@digital.justice.gov.uk",
							"cloud-platform.justice.gov.uk/slack-channel": "test_channel",
							"cloud-platform.justice.gov.uk/source-code":   "https://github.com/ministryofjustice/testrepo.git",
							"cloud-platform.justice.gov.uk/team-name":     "test-team",
						},
					},
				},
			},
			want: namespace.Namespace{
				Name:             "ns1",
				Application:      "test-app",
				BusinessUnit:     "test-bu",
				DeploymentType:   "test",
				GithubURL:        "https://github.com/ministryofjustice/testrepo.git",
				TeamName:         "test-team",
				TeamSlackChannel: "test_channel",
				DomainNames:      []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNamespaceDetails(tt.args.ns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNamespaceDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildIngressesMap(t *testing.T) {
	type args struct {
		ingressItems []networking.Ingress
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{
			name: "ns1",
			args: args{
				ingressItems: []networking.Ingress{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "ingress-01",
							Namespace: "ns-01",
							Annotations: map[string]string{
								"external-dns.alpha.kubernetes.io/aws-weight":     "true",
								"external-dns.alpha.kubernetes.io/set-identifier": "ingress-01-ns-01-blue",
							},
						},
						Spec: networking.IngressSpec{
							TLS: []networking.IngressTLS{
								{
									Hosts: []string{"example-01.com"},
								},
							},
							Rules: []networking.IngressRule{
								{
									Host: "example-01.com",
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "ingress-02",
							Namespace: "ns-02",
							Annotations: map[string]string{
								"external-dns.alpha.kubernetes.io/aws-weight":     "true",
								"external-dns.alpha.kubernetes.io/set-identifier": "ingress-02-ns-02-blue",
							},
						},
						Spec: networking.IngressSpec{
							TLS: []networking.IngressTLS{
								{
									Hosts: []string{"example-02.com"},
								},
							},
							Rules: []networking.IngressRule{
								{
									Host: "example-02.com",
								},
							},
						},
					},
				},
			},
			want: map[string][]string{
				"ns-01": {"example-01.com"},
				"ns-02": {"example-02.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildIngressesMap(tt.args.ingressItems); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildIngressesMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
