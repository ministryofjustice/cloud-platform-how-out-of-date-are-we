package main

import (
	"reflect"
	"testing"
)

func Test_perdayCount(t *testing.T) {
	type args struct {
		migratedDates []string
	}
	tests := []struct {
		name string
		args args
		want map[string]int
	}{
		{
			name: "Slice of dates should return the number of occurance",
			args: args{
				migratedDates: []string{"2021-11-03", "2021-11-03", "2021-11-03", "2021-11-04"},
			},
			want: map[string]int{
				"2021-11-03": 3,
				"2021-11-04": 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := perdayCount(tt.args.migratedDates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("perdayCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildMigratedSlice(t *testing.T) {
	type args struct {
		nsCountPerDate map[string]int
	}
	tests := []struct {
		name string
		args args
		want []map[string]string
	}{
		{
			name: "Date occurance map return a slice with correct values filled",
			args: args{
				nsCountPerDate: map[string]int{
					"2021-11-03": 3,
					"2021-11-04": 1,
				},
			},
			want: []map[string]string{
				{
					"date":       "2021-11-03",
					"todayCount": "3",
					"tillCount":  "7",
					"percentage": "1.96",
				},
				{
					"date":       "2021-11-04",
					"todayCount": "1",
					"tillCount":  "8",
					"percentage": "2.23",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildMigratedSlice(tt.args.nsCountPerDate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildMigratedSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
