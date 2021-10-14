package obj_utils

import (
	"testing"

	"github.com/xutils/lib-common/xlog"
	"github.com/stretchr/testify/assert"
)

type Obj1 struct {
	F1       int64
	F2       string `json:"f_2"`
	FreeTest string `json:"free_test"`
}

func TestMustCopyMap2EntityWrapper(t *testing.T) {
	w, err := CompileCopyMap2EntityWrapper(Obj1{})
	assert.Nil(t, err)
	xlog.Info("w=%+v", w)

	obj := &Obj1{}
	src := map[string]interface {
	}{
		"f_2": "test",
		"F2":  "TEST",
	}
	skipped, err := w.ConvertMapToEntity(src, obj)
	assert.Equal(t, obj.F2, "TEST")
	xlog.Info("skipped=%v||err=%v||obj=%+v", skipped, err, obj)
	src = map[string]interface {
	}{
		"f_2": "test",
		"F1":  "0",
	}

	skipped, err = w.ConvertMapToEntity(src, obj)
	xlog.Info("skipped=%v||err=%v||obj=%+v", skipped, err, obj)

	assert.NotNil(t, err)
	src = map[string]interface {
	}{
		"f_2": "test",
		"F1":  int64(100),
	}
	obj = &Obj1{
		FreeTest: "tttt",
	}
	skipped, err = w.ConvertMapToEntity(src, obj)
	xlog.Info("skipped=%v||err=%v||obj=%+v", skipped, err, obj)

	assert.Nil(t, err)
	assert.Equal(t, "tttt", obj.FreeTest)
	assert.Equal(t, "test", obj.F2)
	assert.Equal(t, int64(100), obj.F1)

}
