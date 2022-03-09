package main

import (
	"reflect"
	"testing"
)

func Test_costsByNamespace(t *testing.T) {
	type args struct {
		awsCostUsageData [][]string
	}
	tests := []struct {
		name string
		args args
		want map[string]map[string]float64
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
			want: map[string]map[string]float64{
				"ns1": map[string]float64{
					"service 1": 16.40,
					"service 2": 1.40,
				},
				"ns2": map[string]float64{
					"service 2": 12.90,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := costsByNamespace(tt.args.awsCostUsageData); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("costsByNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
