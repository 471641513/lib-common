package obj_utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/opay-org/lib-common/xlog"
)

/*
func convertEntityToMap(data interface{}, tagList []string) (output map[string]interface{}) {
	output = map[string]interface{}{}
	v := reflect.ValueOf(data).Elem()

	tagListMap := map[string]interface{}{}

	for _, tag := range tagList {
		tagListMap[tag] = nil
	}

	for i := 0; i < v.NumField(); i++ {
		fieldTag := v.Type().Field(i).Tag.Get("gorm")
		if fieldTag == "-" {
			continue
		}
		fieldTag = strings.Replace(fieldTag, "column:", "", 1)

		if _, ok := tagListMap[fieldTag]; ok {
			output[fieldTag] = v.Field(i).Interface()
		}

	}
	return
}

*/

type CopyEntity2MapWrapper struct {
	srcType reflect.Type

	gormField2idx map[string]int
}

func (w *CopyEntity2MapWrapper) CopyEntity2MapWrapper() *CopyEntity2MapWrapper {
	return w
}

func MustCompileCopyEntity2MapWrapper(src interface{}) (w *CopyEntity2MapWrapper) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	if src == nil {
		err = fmt.Errorf("src or dest nil||src=%v", src)
		return
	}

	if reflect.TypeOf(src).Kind() != reflect.Struct {
		xlog.Warn("reflect.TypeOf(src).Kind() =%v", reflect.TypeOf(src).Kind())
		err = fmt.Errorf("src not struct")
		return
	}
	w, err = compileCopyEntity2MapWrapper(reflect.TypeOf(src))
	return
}

func CompileCopyEntity2MapWrapper(src interface{}) (w *CopyEntity2MapWrapper, err error) {
	if src == nil {
		err = fmt.Errorf("src or dest nil||src=%v", src)
		return
	}

	if reflect.TypeOf(src).Kind() != reflect.Struct {
		xlog.Warn("reflect.TypeOf(src).Kind() =%v", reflect.TypeOf(src).Kind())
		err = fmt.Errorf("src not struct")
		return
	}
	return compileCopyEntity2MapWrapper(reflect.TypeOf(src))
}

func compileCopyEntity2MapWrapper(srcType reflect.Type) (w *CopyEntity2MapWrapper, err error) {
	w = &CopyEntity2MapWrapper{
		srcType:       srcType,
		gormField2idx: map[string]int{},
	}

	for i := 0; i < srcType.NumField(); i++ {
		fieldTag := srcType.Field(i).Tag.Get("gorm")
		if fieldTag == "-" {
			continue
		}
		fieldTag = strings.Replace(fieldTag, "column:", "", 1)
		w.gormField2idx[fieldTag] = i
	}
	return
}

func (w *CopyEntity2MapWrapper) ConvertEntityToMap(data interface{}, fields []string) (output map[string]interface{}, skipfields []string, err error) {

	output = map[string]interface{}{}

	var val reflect.Value
	if reflect.TypeOf(data).Kind() == reflect.Ptr {
		val = reflect.ValueOf(data).Elem()
	} else {
		val = reflect.ValueOf(data)
	}

	if val.Type() != w.srcType {
		err = fmt.Errorf("illegal data type||want=(%v) get=(%v)", w.srcType, val.Type())
		return
	}

	for _, field := range fields {
		if idx, exists := w.gormField2idx[field]; exists {
			output[field] = val.Field(idx).Interface()
		} else {
			skipfields = append(skipfields, field)
		}
	}
	return
}
