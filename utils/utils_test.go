package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/opay-org/lib-common/xlog"
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

func TestEarthDistance(t *testing.T) {
	type args struct {
		lat1 float64
		lng1 float64
		lat2 float64
		lng2 float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			args: args{
				lat1: 6.505989,
				lng1: 3.392925,
				lat2: 6.505989,
				lng2: 3.392925,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EarthDistance(tt.args.lat1, tt.args.lng1, tt.args.lat2, tt.args.lng2); got != tt.want {
				t.Errorf("EarthDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
