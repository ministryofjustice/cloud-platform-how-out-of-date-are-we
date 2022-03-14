package main

import (
	"testing"
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
