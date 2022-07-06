package main

import (
	"reflect"
	"testing"
)

func Test_getFirstLastDayofMonth(t *testing.T) {
	type args struct {
		nthMonth int
	}
	tests := []struct {
		name string
		args args
		want date
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFirstLastDayofMonth(tt.args.nthMonth); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFirstLastDayofMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}
