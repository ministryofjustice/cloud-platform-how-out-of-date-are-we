package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"k8s.io/api/networking/v1beta1"
	networking "k8s.io/api/networking/v1beta1"
)

func TestIngressWithoutAnnotation(t *testing.T) {

	type args struct {
		ingressList *v1beta1.IngressList
	}
	tests := []struct {
		name          string
		args          args
		shouldContain string
	}{
		{
			name: "Both annotations exists",
			args: args{
				ingressList: &networking.IngressList{
					Items: []networking.Ingress{
						{
							ObjectMeta: v1.ObjectMeta{
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
					},
				},
			},
			shouldContain: `{
				{
				"updated_at":        `+time.Now().Format("2006-01-2 15:4:5 UTC")+`,
				"weighting_ingress": [],
				}`,
			},
		},
		{
			name: "Only aws-weight annotation exists",
			args: args{
				ingressList: &networking.IngressList{
					Items: []networking.Ingress{
						{
							ObjectMeta: v1.ObjectMeta{
								Name:      "ingress-02",
								Namespace: "ns-02",
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
			},
			shouldContain: `{
				{
				"updated_at":        `+time.Now().Format("2006-01-2 15:4:5 UTC")+`,
				"weighting_ingress": [],
				}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IngressWithoutAnnotation(tt.args.ingressList)
			fmt.Printf("got = %s, want = %s", string(got), tt.shouldContain)
			if (err != nil) && strings.ContainsAny(string(got), tt.shouldContain) {
				t.Errorf("IngressWithoutAnnotation() error = %v, got = %v, wantErr %v", err, string(got), tt.shouldContain)
				return
			}
		})
	}

}
