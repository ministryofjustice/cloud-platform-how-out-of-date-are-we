package main

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestGetPodResourceDetails(t *testing.T) {
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name               string
		args               args
		wantR              NamespaceResource
		wantNamespace      string
		wantContainerCount int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotNamespace, gotContainerCount := GetPodResourceDetails(tt.args.pod)
			if !reflect.DeepEqual(gotR, tt.wantR) {
				t.Errorf("GetPodResourceDetails() gotR = %v, want %v", gotR, tt.wantR)
			}
			if gotNamespace != tt.wantNamespace {
				t.Errorf("GetPodResourceDetails() gotNamespace = %v, want %v", gotNamespace, tt.wantNamespace)
			}
			if gotContainerCount != tt.wantContainerCount {
				t.Errorf("GetPodResourceDetails() gotContainerCount = %v, want %v", gotContainerCount, tt.wantContainerCount)
			}
		})
	}
}

func TestGetPodUsageDetails(t *testing.T) {
	type args struct {
		PodMetrics v1beta1.PodMetrics
	}
	tests := []struct {
		name          string
		args          args
		wantU         NamespaceResource
		wantNamespace string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotU, gotNamespace := GetPodUsageDetails(tt.args.PodMetrics)
			if !reflect.DeepEqual(gotU, tt.wantU) {
				t.Errorf("GetPodUsageDetails() gotU = %v, want %v", gotU, tt.wantU)
			}
			if gotNamespace != tt.wantNamespace {
				t.Errorf("GetPodUsageDetails() gotNamespace = %v, want %v", gotNamespace, tt.wantNamespace)
			}
		})
	}
}

func TestGetPodHardLimits(t *testing.T) {
	type args struct {
		resourceQuota corev1.ResourceQuota
	}
	tests := []struct {
		name          string
		args          args
		wantH         NamespaceResource
		wantNamespace string
		wantErr       bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotH, gotNamespace, err := GetPodHardLimits(tt.args.resourceQuota)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodHardLimits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotH, tt.wantH) {
				t.Errorf("GetPodHardLimits() gotH = %v, want %v", gotH, tt.wantH)
			}
			if gotNamespace != tt.wantNamespace {
				t.Errorf("GetPodHardLimits() gotNamespace = %v, want %v", gotNamespace, tt.wantNamespace)
			}
		})
	}
}

func Test_addResourceList(t *testing.T) {
	type args struct {
		list corev1.ResourceList
		new  corev1.ResourceList
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addResourceList(tt.args.list, tt.args.new)
		})
	}
}
