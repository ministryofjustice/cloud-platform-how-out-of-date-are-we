package hoodaw

import (
	"testing"
)

func TestQueryApi(t *testing.T) {
	type args struct {
		endPoint string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test correct endpoint",
			args: args{
				endPoint: "ingress-weighting",
			},
			wantErr: false,
		},
		{
			name: "Test incorrect endpoint",
			args: args{
				endPoint: "%",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := QueryApi(tt.args.endPoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryApi() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
