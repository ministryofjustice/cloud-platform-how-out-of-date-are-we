package main

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_costs_updatecostsByNamespace(t *testing.T) {
	type fields struct {
		costPerNamespace map[string]map[string]float64
	}
	type args struct {
		awsCostUsageData [][]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "single slice of array return a map",
			args: args{
				awsCostUsageData: [][]string{
					{"20221234", "service 1", "ns1", "16.40"},
					{"20221234", "service 2", "ns2", "11.70"},
					{"20221234", "service 2", "ns1", "1.40"},
					{"20221234", "service 2", "ns2", "1.20"},
				},
			},
			fields: fields{
				costPerNamespace: map[string]map[string]float64{
					"ns1": map[string]float64{
						"service 1": 16.40,
						"service 2": 1.40,
					},
					"ns2": map[string]float64{
						"service 2": 12.90,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &costs{
				costPerNamespace: tt.fields.costPerNamespace,
			}
			if err := c.updatecostsByNamespace(tt.args.awsCostUsageData); (err != nil) != tt.wantErr {
				t.Errorf("costs.updatecostsByNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_costs_getSharedCosts(t *testing.T) {
	type fields struct {
		costPerNamespace map[string]map[string]float64
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{

			name: "get shared costs per namespace",
			fields: fields{
				costPerNamespace: map[string]map[string]float64{
					"SHARED_COSTS": map[string]float64{
						"service 1": 36,
						"service 2": 42,
						"service 3": 21,
					},
					"ns1": map[string]float64{
						"service 3": 12.90,
					},
					"ns2": map[string]float64{
						"service 1": 12.90,
					},
					"ns3": map[string]float64{
						"service 2": 12.90,
					},
					"ns4": map[string]float64{
						"service 1": 12.90,
					},
				},
			},
			want: 24.75,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &costs{
				costPerNamespace: tt.fields.costPerNamespace,
			}
			if got := c.getSharedCosts(); got != tt.want {
				t.Errorf("costs.updateSharedCosts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_costs_buildCostsResourceMap(t *testing.T) {
	type fields struct {
		costPerNamespace map[string]map[string]float64
	}
	type args struct {
		nsList []v1.Namespace
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   resourceMap
	}{
		{
			name: "for given costPerNamespace build resourceMap",
			fields: fields{
				costPerNamespace: map[string]map[string]float64{
					"ns1": map[string]float64{
						"service 1": 16.40,
						"service 2": 1.40,
					},
					"ns2": map[string]float64{
						"service 2": 12.90,
					},
				},
			},
			args: args{
				nsList: []v1.Namespace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ns1",
						},
					},
				},
			},
			want: resourceMap{
				"ns1": resourceMap{
					"breakdown": map[string]float64{
						"service 1": 16.40,
						"service 2": 1.40,
					},
					"total": 17.80,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &costs{
				costPerNamespace: tt.fields.costPerNamespace,
			}
			if got := c.buildCostsResourceMap(tt.args.nsList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("costs.buildCostsResourceMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_costs_addResource(t *testing.T) {
	type fields struct {
		costPerNamespace map[string]map[string]float64
	}
	type args struct {
		ns       string
		resource string
		cost     float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name: "when one service and cost given, add resource map to ns",
			fields: fields{
				costPerNamespace: map[string]map[string]float64{
					"ns1": map[string]float64{
						"service 1": 16.40,
					},
				},
			},
			args: args{
				ns:       "ns1",
				resource: "service 2",
				cost:     1.7,
			},
			want: fields{
				costPerNamespace: map[string]map[string]float64{
					"ns1": map[string]float64{
						"service 1": 16.40,
						"service 2": 1.7,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &costs{
				costPerNamespace: tt.fields.costPerNamespace,
			}
			c.addResource(tt.args.ns, tt.args.resource, tt.args.cost)
		})
	}
}
