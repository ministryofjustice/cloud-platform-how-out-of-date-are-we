package main

import (
	"reflect"
	"testing"

	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func Test_live1DomainSearch(t *testing.T) {
	type args struct {
		domainSearch *v1beta1.IngressList
	}
	tests := []struct {
		name    string
		args    args
		want    []map[string]string
		wantErr bool
	}{
		{
			name: "live1DomainSearch-Success",
			args: args{
				domainSearch: &v1beta1.IngressList{
					Items: []v1beta1.Ingress{
						{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "namespace-1",
								Name:      "ingress-1",
							},
							Spec: v1beta1.IngressSpec{
								TLS: []v1beta1.IngressTLS{
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
					"hostname":  "example.live-1.cloud-platform.service.justice.gov.uk",
					"namespace": "namespace-1",
					"ingress":   "ingress-1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := live1DomainSearch(tt.args.domainSearch)
			if (err != nil) != tt.wantErr {
				t.Errorf("live1DomainSearch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("live1DomainSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}
