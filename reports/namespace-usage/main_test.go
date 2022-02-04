package main

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestgetPodResourceDetails(t *testing.T) {
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
		{
			name: "Pod with resource requests",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse("1"),
									"memory": resource.MustParse("100Mi"),
								},
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse("10"),
									"memory": resource.MustParse("1000Mi"),
								},
							}},
						},
					},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{Type: v1.PodInitialized, Status: v1.ConditionTrue},
						},
					},
				},
			},
			wantR: NamespaceResource{
				CPU:    1000,
				Memory: 100,
				Pods:   0,
			},
			wantNamespace:      "test",
			wantContainerCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotNamespace, gotContainerCount := getPodResourceDetails(tt.args.pod)
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

func TestgetPodUsageDetails(t *testing.T) {
	type args struct {
		podMetrics v1beta1.PodMetrics
	}
	tests := []struct {
		name          string
		args          args
		wantU         NamespaceResource
		wantNamespace string
	}{
		{
			name: "Pod with resource Metrics",
			args: args{
				podMetrics: v1beta1.PodMetrics{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Containers: []v1beta1.ContainerMetrics{
						{
							Name: "app1",
							Usage: v1.ResourceList{
								"cpu":    resource.MustParse("1"),
								"memory": resource.MustParse("200Mi"),
							},
						},
					},
				},
			},
			wantU: NamespaceResource{
				CPU:    1000,
				Memory: 200,
				Pods:   0,
			},
			wantNamespace: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotU, gotNamespace := getPodUsageDetails(tt.args.podMetrics)
			if !reflect.DeepEqual(gotU, tt.wantU) {
				t.Errorf("GetPodUsageDetails() gotU = %v, want %v", gotU, tt.wantU)
			}
			if gotNamespace != tt.wantNamespace {
				t.Errorf("GetPodUsageDetails() gotNamespace = %v, want %v", gotNamespace, tt.wantNamespace)
			}
		})
	}
}

func TestgetPodHardLimits(t *testing.T) {
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
		{
			name: "resourcequota for given namespace",
			args: args{
				resourceQuota: corev1.ResourceQuota{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Status: corev1.ResourceQuotaStatus{
						Hard: v1.ResourceList{
							"pods": resource.MustParse("50"),
						},
						Used: v1.ResourceList{
							"pods": resource.MustParse("2"),
						},
					},
				},
			},
			wantH: NamespaceResource{
				CPU:    0,
				Memory: 0,
				Pods:   50,
			},
			wantNamespace: "test",
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotH, gotNamespace, err := getPodHardLimits(tt.args.resourceQuota)
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
		{
			name: "add resourceList",
			args: args{
				list: corev1.ResourceList{
					"cpu":    resource.MustParse("1"),
					"memory": resource.MustParse("200Mi"),
				},
				new: corev1.ResourceList{
					"cpu":    resource.MustParse("1"),
					"memory": resource.MustParse("100Mi"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addResourceList(tt.args.list, tt.args.new)
		})
	}
}

func Test_getAllPodResourceDetails(t *testing.T) {
	type args struct {
		PodList *v1.PodList
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]NamespaceResource
		want1   map[string]int
		wantErr bool
	}{
		{
			name: "get resource requests for a given clientset",
			args: args{
				PodList: &v1.PodList{
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "pod-01",
								Namespace: "ns-01",
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											"cpu":    resource.MustParse("1"),
											"memory": resource.MustParse("100Mi"),
										},
										Limits: v1.ResourceList{
											"cpu":    resource.MustParse("10"),
											"memory": resource.MustParse("1000Mi"),
										},
									}},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "pod-02",
								Namespace: "ns-02",
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											"cpu":    resource.MustParse("2"),
											"memory": resource.MustParse("200Mi"),
										},
									}},
								},
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "pod-03",
								Namespace: "ns-03",
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{Resources: v1.ResourceRequirements{
										Requests: v1.ResourceList{
											"cpu":    resource.MustParse("3"),
											"memory": resource.MustParse("300Mi"),
										},
									}},
								},
							},
						},
					},
				},
			},
			want: map[string]NamespaceResource{
				"ns-01": NamespaceResource{
					CPU:    1000,
					Memory: 100,
					Pods:   0,
				},
				"ns-02": NamespaceResource{
					CPU:    2000,
					Memory: 200,
					Pods:   0,
				},
				"ns-03": NamespaceResource{
					CPU:    3000,
					Memory: 300,
					Pods:   0,
				},
			},
			want1: map[string]int{
				"ns-01": 1,
				"ns-02": 1,
				"ns-03": 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		kubeClient := testclient.NewSimpleClientset(tt.args.PodList)
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getAllPodResourceDetails(kubeClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAllPodResourceDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllPodResourceDetails() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getAllPodResourceDetails() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNamespaceResource_addNamespaceResource(t *testing.T) {
	type fields struct {
		CPU    float64
		Memory float64
		Pods   int
	}
	type args struct {
		new NamespaceResource
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := &NamespaceResource{
				CPU:    tt.fields.CPU,
				Memory: tt.fields.Memory,
				Pods:   tt.fields.Pods,
			}
			list.addNamespaceResource(tt.args.new)
		})
	}
}
