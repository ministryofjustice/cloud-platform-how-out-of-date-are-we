package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "Absolute path to the kubeconfig file.")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file.")
	}
	flag.Parse()

	// use the current context in kubeconfig

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Println(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println(err.Error())
	}

	ingress, err := clientset.NetworkingV1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
	}

	// For each ingress resource, check
	m := make(map[string]v1.IngressTLS)
	for _, i := range ingress.Items {
		if _, ok := i.Annotations["external-dns.alpha.kubernetes.io/aws-weight"]; !ok {
			m[i.GetNamespace()+"/"+i.GetName()] = i.Spec.TLS[0]
		}
	}

	jsonStr, err := json.Marshal(m)
	fmt.Println(string(jsonStr))
}
