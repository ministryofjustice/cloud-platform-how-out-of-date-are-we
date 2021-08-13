package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	hoodawApiKey = flag.String("hoodawAPIKey", os.Getenv("HOODAW_API_KEY"), "API key to post data to the 'How out of date are we' API")
	hoodawHost   = flag.String("hoodawHost", os.Getenv("HOODAW_HOST"), "Hostname of the 'How out of date are we' API")
)

func main() {
	var kubeconfig *string
	hoodawEndpoint := "/ingress-weighting"

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

	ingress, err := clientset.NetworkingV1beta1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println(err.Error())
	}

	// For each ingress resource, check
	m := make(map[string][]string)
	for _, i := range ingress.Items {
		if _, ok := i.Annotations["external-dns.alpha.kubernetes.io/aws-weight"]; !ok {
			for _, v := range i.Spec.TLS {
				m[i.GetNamespace()+"/"+i.GetName()] = v.Hosts
			}
		}
	}

	jsonStr, err := json.Marshal(m)
	fmt.Println(string(jsonStr))

	req, err := http.NewRequest("POST", *hoodawHost+*&hoodawEndpoint, nil)

	req.Header.Add("X-API-KEY", *hoodawApiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error on POST response. \n", err)
	}

	defer resp.Body.Close()
}
