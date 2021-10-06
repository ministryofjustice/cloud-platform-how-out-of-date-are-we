package main

import (
	"strings"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	networking "k8s.io/api/networking/v1beta1"
)

func TestIngressWithoutAnnotation(t *testing.T) {

	want := "\"weighting_ingress\":[]"
	var ingress_01 = &networking.IngressList{
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
					Rules: []networking.IngressRule{
						{
							Host: "example.com",
						},
					},
				},
			},
		},
	}

	got, err := IngressWithoutAnnotation(ingress_01)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if !strings.Contains(string(got), want) {
		t.Errorf("Unexpected error: %s", got)
	}

}
