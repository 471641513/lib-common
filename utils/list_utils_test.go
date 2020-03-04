package utils

import (
	"reflect"
	"testing"
)

func TestIntListIntersect(t *testing.T) {
	type args struct {
		l1 []int64
		l2 []int64
	}
	tests := []struct {
		name          string
		args          args
		wantIntersect []int64
	}{{
		args: args{
			l1: []int64{1, 2, 3, 4},
			l2: []int64{},
		},
		wantIntersect: []int64{},
	}, {
		args: args{
			l1: []int64{1, 2, 2, 4},
			l2: []int64{2},
		},
		wantIntersect: []int64{2},
	}, {
		args: args{
			l1: []int64{1, 2, 3, 4},
			l2: []int64{2, 3, 4, 6, 7},
		},
		wantIntersect: []int64{2, 3, 4},
	}, {
		args: args{
			l1: []int64{1, 2, 3, 4},
			l2: []int64{0, 1, 4, 7},
		},
		wantIntersect: []int64{1, 4},
	}, {
		args: args{
			l1: []int64{1, 2, 3, 4},
			l2: []int64{6, 7, 8},
		},
		wantIntersect: []int64{},
	}, {
		args: args{
			l1: []int64{1, 2, 2, 3, 4},
			l2: []int64{1, 2, 2, 3, 4},
		},
		wantIntersect: []int64{1, 2, 2, 3, 4},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIntersect := IntListIntersect(tt.args.l1, tt.args.l2); len(gotIntersect) == 0 && len(tt.wantIntersect) == 0 {
			} else {
				if !reflect.DeepEqual(gotIntersect, tt.wantIntersect) {
					t.Errorf("IntListIntersect() = %v, want %v", gotIntersect, tt.wantIntersect)
				}
			}
			if gotIntersect := IntListIntersect(tt.args.l2, tt.args.l1); len(gotIntersect) == 0 && len(tt.wantIntersect) == 0 {
			} else {
				if !reflect.DeepEqual(gotIntersect, tt.wantIntersect) {
					t.Errorf("IntListIntersect() = %v, want %v", gotIntersect, tt.wantIntersect)
				}
			}
		})
	}
}
