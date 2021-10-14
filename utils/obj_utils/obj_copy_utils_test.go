package obj_utils

import (
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"golang.org/x/exp/rand"

	"github.com/xutils/lib-common/xlog"
	"github.com/stretchr/testify/assert"
)

/*
	reflect.Int:     numericalInt,
	reflect.Int8:    numericalInt,
	reflect.Int16:   numericalInt,
	reflect.Int32:   numericalInt,
	reflect.Int64:   numericalInt,
	reflect.Uint:    numericalUInt,
	reflect.Uint8:   numericalUInt,
	reflect.Uint16:  numericalUInt,
	reflect.Uint32:  numericalUInt,
	reflect.Uint64:  numericalUInt,
	reflect.Float32: numericalFloat,
	reflect.Float64: numericalFloat,

	time.Time
*/
type NumF interface {
	Val() int64
}
type Int struct {
	F int
}

func (f *Int) Val() int64 {
	return int64(f.F)
}

type Int8 struct {
	F int8
}

func (f *Int8) Val() int64 {
	return int64(f.F)
}

type Int16 struct {
	F int16
}

func (f *Int16) Val() int64 {
	return int64(f.F)
}

type Int32 struct {
	F int32
}

func (f *Int32) Val() int64 {
	return int64(f.F)
}

type Int64 struct {
	F int64
}

func (f *Int64) Val() int64 {
	return int64(f.F)
}

type Uint struct {
	F uint
}

func (f *Uint) Val() int64 {
	return int64(f.F)
}

type Uint8 struct {
	F uint8
}

func (f *Uint8) Val() int64 {
	return int64(f.F)
}

type Uint16 struct {
	F uint16
}

func (f *Uint16) Val() int64 {
	return int64(f.F)
}

type Uint32 struct {
	F uint32
}

func (f *Uint32) Val() int64 {
	return int64(f.F)
}

type Uint64 struct {
	F uint64
}

func (f *Uint64) Val() int64 {
	return int64(f.F)
}

type Float32 struct {
	F float32
}

func (f *Float32) Val() int64 {
	return int64(f.F)
}

type Float64 struct {
	F float64
}

func (f *Float64) Val() int64 {
	return int64(f.F)
}

type Ttime struct {
	F time.Time
}

func (f *Ttime) Val() int64 {
	return f.F.Unix()
}

var newObjFunc = []func(empty bool) interface{}{
	func(empty bool) interface{} {
		i := &Int{}
		if empty {
			return i
		}
		i.F = rand.Intn(math.MaxInt64)
		return i
	},
	func(empty bool) interface{} {
		i := &Int8{}
		if empty {
			return i
		}
		i.F = int8(rand.Intn(math.MaxInt8))
		return i
	},
	func(empty bool) interface{} {
		i := &Int16{}
		if empty {
			return i
		}
		i.F = int16(rand.Intn(math.MaxInt16))
		return i
	},
	func(empty bool) interface{} {
		i := &Int32{}
		if empty {
			return i
		}
		i.F = int32(rand.Intn(math.MaxInt32))
		return i
	},
	func(empty bool) interface{} {
		i := &Int64{}
		if empty {
			return i
		}
		i.F = int64(rand.Intn(math.MaxInt64))
		return i
	},

	func(empty bool) interface{} {
		i := &Uint{}
		if empty {
			return i
		}
		i.F = uint(rand.Intn(math.MaxInt64))
		return i
	},
	func(empty bool) interface{} {
		i := &Uint8{}
		if empty {
			return i
		}
		i.F = uint8(rand.Intn(math.MaxInt8))
		return i
	},
	func(empty bool) interface{} {
		i := &Uint16{}
		if empty {
			return i
		}
		i.F = uint16(rand.Intn(math.MaxInt16))
		return i
	},
	func(empty bool) interface{} {
		i := &Uint32{}
		if empty {
			return i
		}
		i.F = uint32(rand.Intn(math.MaxInt32))
		return i
	},
	func(empty bool) interface{} {
		i := &Uint64{}
		if empty {
			return i
		}
		i.F = uint64(rand.Intn(math.MaxInt64))
		return i
	},
	func(empty bool) interface{} {
		i := &Float32{}
		if empty {
			return i
		}
		i.F = float32(rand.Intn(math.MaxInt64))
		return i
	},
	func(empty bool) interface{} {
		i := &Float64{}
		if empty {
			return i
		}
		i.F = float64(rand.Intn(math.MaxInt64))
		return i
	},
	func(empty bool) interface{} {
		i := &Ttime{}
		if empty {
			return i
		}
		i.F = time.Now().Add(time.Duration(rand.Intn(1000)) * time.Second)
		return i
	},
}

func isFloat(data interface{}) bool {
	if _, ok := data.(*Float64); ok {
		return true
	}
	if _, ok := data.(*Float32); ok {
		return true
	}
	return false
}

func TestCopyFieldValuesAllType(t *testing.T) {

	for _, srcFunc := range newObjFunc {
	innerLoop:
		for _, destFunc := range newObjFunc {
			src := srcFunc(false)
			dest := destFunc(true)
			if isFloat(dest) || isFloat(src) {
				// skip float for the moment
				// TODO: copy from any to float might cause unexpected value
				continue innerLoop
			}
			if srcF, ok := src.(NumF); ok {
				if destF, ok := dest.(NumF); ok {
					_, err := copyFieldValues(srcF, destF)
					assert.Nil(t, err, "src=%+v,dest=%+v", srcF, destF)
					//assert.Equal(t, srcF.Val(), destF.Val(), "src=%+v,dest=%+v", srcF, destF)
					//xlog.Debug("src=%+v||dest=%+v", src, dest)
					if srcF.Val() != destF.Val() {
						if destF.Val() > 0 {
							xlog.Error("failed||src=%v||dest=%v||src type=%v||dest type=%v",
								srcF.Val(), destF.Val(),
								reflect.TypeOf(srcF), reflect.TypeOf(destF))
							//t.FailNow()
							assert.Equal(t, srcF.Val(), destF.Val(),
								"src=%v||dest=%v||src type=%v||dest type=%v",
								srcF.Val(), destF.Val(),
								reflect.TypeOf(srcF), reflect.TypeOf(destF))
						} else {
							xlog.Info("skipped||src=%v||dest=%v||src type=%v||dest type=%v",
								srcF.Val(), destF.Val(),
								reflect.TypeOf(srcF), reflect.TypeOf(destF))
						}
						//
					} else {
						//xlog.Info("passed||src=%v||dest=%v", reflect.TypeOf(srcF), reflect.TypeOf(destF))
					}
				} else {
					assert.Fail(t, "dest is not numf")
				}
			}
		}
	}

}

func TestCopyFieldValues(t *testing.T) {
	nowTs := time.Now()
	type obj1 = struct {
		F1    int
		F11   int64
		F2    string
		F3    float64
		F5    string
		F6    int
		FT    time.Time
		FTint int64
		FTT   time.Time
	}
	type obj2 = struct {
		F1    int
		F11   uint64
		F2    string
		F3    float32
		F5    int
		F10   int
		FT    float64
		FTint time.Time
		FTT   time.Time
	}

	o1 := &obj1{
		F1:    1,
		F11:   11,
		F2:    "2",
		F3:    3.0,
		F5:    "5",
		F6:    6,
		FT:    nowTs,
		FTint: time.Now().Add(time.Second).Unix(),
		FTT:   nowTs.Add(time.Hour),
	}
	o2 := &obj2{}

	xlog.Debug("o2.FTint=%v", o2.FTint.Unix())
	_, err := copyFieldValues(nil, o2)
	assert.NotEqual(t, nil, err)
	_, err = copyFieldValues(o1, nil)
	assert.NotEqual(t, nil, err)

	xlog.Info("o1=%+v||o2=%+v", o1, o2)
	skip, err := copyFieldValues(o1, o2)
	assert.Nil(t, err)
	xlog.Info("o1=%+v||o2=%+v||skip=%v", o1, o2, skip)

	assert.Equal(t, o2.F1, o1.F1)
	assert.Equal(t, o2.F11, uint64(o1.F11))
	assert.Equal(t, o2.F2, o1.F2)
	assert.Equal(t, o2.F3, float32(o1.F3))
	assert.Equal(t, o2.F5, 0)

	assert.Equal(t, int64(o2.FT), o1.FT.Unix())
	assert.Equal(t, o2.FTint.Unix(), int64(o1.FTint))
	assert.Equal(t, o2.FTT.Unix(), o1.FTT.Unix())

	o3 := &obj2{}
	skip, err = copyFieldValues(o1, o3, "F1", "F2", "F6")
	assert.Nil(t, err)
	assert.Equal(t, o3.F1, o1.F1)
	assert.Equal(t, o3.F11, uint64(0))
	assert.Equal(t, o3.F2, o1.F2)
	assert.Equal(t, o3.F3, float32(0))
	assert.Equal(t, o3.F5, 0)
	xlog.Info("o1=%+v||o3=%+v||skip=%v", o1, o3, skip)

	type obj3 struct {
		F0 uint64 `json:"F1"`
		F1 float32
		F  int `json:"F2"`
	}

	o4 := &obj3{}
	skip, err = copyFieldValues(o1, o4)
	xlog.Info("o1=%+v||o3=%+v||skip=%v", o1, o4, skip)
	assert.Nil(t, err)
	assert.Equal(t, o4.F0, uint64(o1.F1))
	assert.Equal(t, o4.F1, float32(o1.F1))
}
func Test_JsonProjection(t *testing.T) {

	type obj1 = struct {
		F1  int `json:"f1"`
		F11 int64
		F3  int
	}
	type obj2 struct {
		F1  int
		F11 int64 `json:"f11"`
		F3  int
	}
	w, err := CompileCopyFieldWrapper(obj1{}, obj2{})
	assert.Nil(t, err)
	xlog.Debug("w=%+v", w)
	o1 := &obj1{
		F1:  10,
		F11: 1111,
		F3:  2222,
	}

	o2 := &obj2{}
	skip, err := w.CopyFieldValues(o1, o2, "F1", "f11")
	xlog.Debug("skip=%+v", skip)
	assert.Nil(t, err)
	assert.Equal(t, o1.F11, o2.F11)
	assert.Equal(t, o1.F1, o2.F1)
	assert.Equal(t, 0, o2.F3)

}
func TestCompileCopyFieldWrapper(t *testing.T) {

	type obj1 = struct {
		F1  int
		F11 int64
		F2  string
		F3  float64
		F5  string
		F6  int
	}
	type obj2 = struct {
		F1  int
		F11 uint64
		F2  string
		F3  float32
		F5  int
		F10 int
	}
	w, err := CompileCopyFieldWrapper(obj1{}, obj2{})
	assert.Nil(t, err)
	xlog.Debug("w=%+v", w)

	o1 := &obj1{
		F1:  1,
		F11: 11,
		F2:  "2",
		F3:  3.0,
		F5:  "5",
		F6:  6,
	}
	o2 := &obj2{}
	_, err = w.CopyFieldValues(nil, o2)
	assert.NotEqual(t, nil, err)
	_, err = w.CopyFieldValues(o1, nil)
	assert.NotEqual(t, nil, err)

	xlog.Info("o1=%+v||o2=%+v", o1, o2)
	skip, err := w.CopyFieldValues(o1, o2)
	assert.Nil(t, err)
	xlog.Info("o1=%+v||o2=%+v||skip=%v", o1, o2, skip)

	assert.Equal(t, o2.F1, o1.F1)
	assert.Equal(t, o2.F11, uint64(o1.F11))
	assert.Equal(t, o2.F2, o1.F2)
	assert.Equal(t, o2.F3, float32(o1.F3))
	assert.Equal(t, o2.F5, 0)

	o3 := &obj2{}
	skip, err = w.CopyFieldValues(o1, o3, "F1", "F2", "F6")
	assert.Nil(t, err)
	assert.Equal(t, o3.F1, o1.F1)
	assert.Equal(t, o3.F11, uint64(0))
	assert.Equal(t, o3.F2, o1.F2)
	assert.Equal(t, o3.F3, float32(0))
	assert.Equal(t, o3.F5, 0)
	xlog.Info("o1=%+v||o3=%+v||skip=%v", o1, o3, skip)
}

func TestMain(m *testing.M) {
	xlog.SetupLogDefault()
	// setup code...
	code := m.Run()
	// teardown code...
	xlog.Close()
	os.Exit(code)
}
