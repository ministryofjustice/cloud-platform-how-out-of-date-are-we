package main

import (
	"reflect"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func Test_live1DomainSearch(t *testing.T) {
	type args struct {
		domainSearch *networkingv1.IngressList
	}
	tests := []struct {
		name string
		args args
		want []map[string]string
	}{
		{
			name: "live1DomainSearch-Success",
			args: args{
				domainSearch: &networkingv1.IngressList{
					Items: []networkingv1.Ingress{
						{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "namespace-1",
								Name:      "ingress-1",
							},
							Spec: networkingv1.IngressSpec{
								TLS: []networkingv1.IngressTLS{
									{
										Hosts: []string{"example.live-1.cloud-platform.service.justice.gov.uk"},
									},
								},
							},
						},
					},
				},
			},
			want: []map[string]string{
				{
					"CreatedAt": "0001-01-1 00:0:0 UTC",
					"hostname":  "example.live-1.cloud-platform.service.justice.gov.uk",
					"namespace": "namespace-1",
					"ingress":   "ingress-1",
				},
			},
		},
		{
			name: "live1DomainSearch-Error",
			args: args{
				domainSearch: &networkingv1.IngressList{
					Items: []networkingv1.Ingress{
						{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "namespace-1",
								Name:      "ingress-1",
							},
							Spec: networkingv1.IngressSpec{
								TLS: []networkingv1.IngressTLS{
									{
										Hosts: []string{"example.live.cloud-platform.service.justice.gov.uk"},
									},
								},
							},
						},
					},
				},
			},
			want: []map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := live1DomainSearch(tt.args.domainSearch)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("live1DomainSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}
