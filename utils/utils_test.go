package utils

import (
	"fmt"
	"github.com/opay-org/lib-common/xlog"
	"testing"
	"time"
)

func TestMustInt(t *testing.T) {
	xlog.SetupLogDefault()
	defer xlog.Close()
	type args struct {
		data interface{}
	}
	tests := []struct {
		name  string
		args  args
		wantI int
	}{
		{
			name: "float64",
			args: args{
				data: float64(1.000),
			},
			wantI: 1,
		},
		{
			name: "float32",
			args: args{
				data: float32(2.000),
			},
			wantI: 2,
		}, {
			name: "string",
			args: args{
				data: "3",
			},
			wantI: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotI := MustInt(tt.args.data); gotI != tt.wantI {
				t.Errorf("MustInt() = %v, want %v", gotI, tt.wantI)
			}
		})
	}
}

func TestGetDayStartTs(t *testing.T) {
	ts := time.Now().Unix()

	gotTs := GetDayStartTs(ts)
	fmt.Printf("%v", gotTs)
}
