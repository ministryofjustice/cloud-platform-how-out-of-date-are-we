package main

import (
	"reflect"
	"testing"
)

func Test_getNamespaces(t *testing.T) {
	type args struct {
		helmListJson string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "helm releases in namespaces",
			args: args{
				helmListJson: "[{\"namespace\":\"ns1\"},{\"namespace\":\"ns2\"}]",
			},
			want:    []string{"ns1", "ns2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNamespaces(tt.args.helmListJson)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}
