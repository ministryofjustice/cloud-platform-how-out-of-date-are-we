package main

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	assert "github.com/stretchr/testify/assert"
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
		shouldContain []map[string]string
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
			shouldContain: []map[string]string{},
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
			shouldContain: []map[string]string{
				{
					"hostname":  "example-02.com",
					"namespace": "ns-02",
					"resource":  "ingress-02",
				},
			},
		},
		{
			name: "Only set-identifier annotation exists",
			args: args{
				ingressList: &networking.IngressList{
					Items: []networking.Ingress{
						{
							ObjectMeta: v1.ObjectMeta{
								Name:      "ingress-03",
								Namespace: "ns-03",
							},
							Spec: networking.IngressSpec{
								TLS: []networking.IngressTLS{
									{
										Hosts: []string{"example-03.com"},
									},
								},
								Rules: []networking.IngressRule{
									{
										Host: "example-03.com",
									},
								},
							},
						},
					},
				},
			},
			shouldContain: []map[string]string{
				{
					"hostname":  "example-03.com",
					"namespace": "ns-03",
					"resource":  "ingress-03",
				},
			},
		},
		{
			name: "Both annotation doesnot exists",
			args: args{
				ingressList: &networking.IngressList{
					Items: []networking.Ingress{
						{
							ObjectMeta: v1.ObjectMeta{
								Name:      "ingress-04",
								Namespace: "ns-04",
							},
							Spec: networking.IngressSpec{
								TLS: []networking.IngressTLS{
									{
										Hosts: []string{"example-04.com"},
									},
								},
								Rules: []networking.IngressRule{
									{
										Host: "example-04.com",
									},
								},
							},
						},
					},
				},
			},
			shouldContain: []map[string]string{
				{
					"hostname":  "example-04.com",
					"namespace": "ns-04",
					"resource":  "ingress-04",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IngressWithoutAnnotation(tt.args.ingressList)
			if err != nil {
				t.Errorf("IngressWithoutAnnotation() error = %v", err)
				return
			}
			if !assert.Equal(t, got, tt.shouldContain) {
				t.Errorf("IngressWithoutAnnotation() got = %v, wantErr %v", got, tt.shouldContain)
			}
		})
	}

}
