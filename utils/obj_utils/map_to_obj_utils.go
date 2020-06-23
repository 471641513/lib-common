package obj_utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/opay-org/lib-common/xlog"
)

type CopyMap2EntityWrapper struct {
	destType       reflect.Type
	field2idx      map[string]int
	fieldsPriority map[string]int
}

func MustCopyMap2EntityWrapper(src interface{}) (w *CopyMap2EntityWrapper) {
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
	w, err = compileCopyMap2EntityWrapper(reflect.TypeOf(src))
	return
}

func CompileCopyMap2EntityWrapper(src interface{}) (w *CopyMap2EntityWrapper, err error) {
	if src == nil {
		err = fmt.Errorf("src or dest nil||src=%v", src)
		return
	}

	if reflect.TypeOf(src).Kind() != reflect.Struct {
		xlog.Warn("reflect.TypeOf(src).Kind() =%v", reflect.TypeOf(src).Kind())
		err = fmt.Errorf("src not struct")
		return
	}
	return compileCopyMap2EntityWrapper(reflect.TypeOf(src))
}

func compileCopyMap2EntityWrapper(destType reflect.Type) (w *CopyMap2EntityWrapper, err error) {
	w = &CopyMap2EntityWrapper{
		destType:       destType,
		field2idx:      map[string]int{},
		fieldsPriority: map[string]int{},
	}

	idx := 0
	for i := 0; i < destType.NumField(); i++ {
		gormfieldTag := destType.Field(i).Tag.Get("gorm")
		if gormfieldTag == "-" {
			continue
		}
		gormfieldTag = strings.Replace(gormfieldTag, "column:", "", 1)

		jsonFieldTag := destType.Field(i).Tag.Get("json")
		if jsonFieldTag == "-" {
			continue
		}

		oriName := destType.Field(i).Name

		if gormfieldTag != "" {
			w.field2idx[gormfieldTag] = i
			w.fieldsPriority[gormfieldTag] = idx
			idx += i
		}

		if jsonFieldTag != "" {
			w.field2idx[jsonFieldTag] = i
			w.fieldsPriority[jsonFieldTag] = idx
			idx += i
		}

		w.field2idx[oriName] = i
		w.fieldsPriority[oriName] = idx
		idx += i
	}
	return
}

func tryCatch(f func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			//xlog.Debug("catch panic | %s\n%s", e, debug.Stack())
			err = fmt.Errorf("%v", e)
		}
	}()
	f()
	return
}

func (w *CopyMap2EntityWrapper) ConvertMapToEntity(src map[string]interface{}, dest interface{}) (skipfields []string, err error) {
	err = tryCatch(func() {
		if reflect.TypeOf(dest).Kind() != reflect.Ptr {
			err = fmt.Errorf("dest is not ptr")
			return
		}
		if dest == nil {
			err = fmt.Errorf("dest ptr is nil")
			return
		}
		var val reflect.Value
		val = reflect.ValueOf(dest).Elem()

		if val.Type() != w.destType {
			err = fmt.Errorf("illegal data type||want=(%v) get=(%v)", w.destType, val.Type())
		}

		fieldsToReplace := map[int]string{}
		for key, _ := range src {
			if idx, ok := w.field2idx[key]; ok {
				priority := w.fieldsPriority[key]
				// otherKey exists
				if otherKey, ok := fieldsToReplace[idx]; ok {
					// compare priority
					otherKeyPriority := w.fieldsPriority[otherKey]
					if otherKeyPriority < priority {
						fieldsToReplace[idx] = key
						skipfields = append(skipfields, otherKey)
					} else {
						skipfields = append(skipfields, key)
					}
				} else {
					fieldsToReplace[idx] = key
				}
			}
		}
		for idx, key := range fieldsToReplace {
			val.Field(idx).Set(reflect.ValueOf(src[key]))
		}
	})
	return
}

//
//type CopyEntity2MapWrapper struct {
//	srcType reflect.Type
//
//	field2idx map[string]int
//}
//
//func (w *CopyEntity2MapWrapper) CopyEntity2MapWrapper() *CopyEntity2MapWrapper {
//	return w
//}
//
//func MustCompileCopyEntity2MapWrapper(src interface{}) (w *CopyEntity2MapWrapper) {
//	var err error
//	defer func() {
//		if err != nil {
//			panic(err)
//		}
//	}()
//	if src == nil {
//		err = fmt.Errorf("src or dest nil||src=%v", src)
//		return
//	}
//
//	if reflect.TypeOf(src).Kind() != reflect.Struct {
//		xlog.Warn("reflect.TypeOf(src).Kind() =%v", reflect.TypeOf(src).Kind())
//		err = fmt.Errorf("src not struct")
//		return
//	}
//	w, err = compileCopyEntity2MapWrapper(reflect.TypeOf(src))
//	return
//}
//
//func CompileCopyEntity2MapWrapper(src interface{}) (w *CopyEntity2MapWrapper, err error) {
//	if src == nil {
//		err = fmt.Errorf("src or dest nil||src=%v", src)
//		return
//	}
//
//	if reflect.TypeOf(src).Kind() != reflect.Struct {
//		xlog.Warn("reflect.TypeOf(src).Kind() =%v", reflect.TypeOf(src).Kind())
//		err = fmt.Errorf("src not struct")
//		return
//	}
//	return compileCopyEntity2MapWrapper(reflect.TypeOf(src))
//}
//
//func compileCopyEntity2MapWrapper(srcType reflect.Type) (w *CopyEntity2MapWrapper, err error) {
//	w = &CopyEntity2MapWrapper{
//		srcType:       srcType,
//		field2idx: map[string]int{},
//	}
//
//	for i := 0; i < srcType.NumField(); i++ {
//		fieldTag := srcType.Field(i).Tag.Get("gorm")
//		if fieldTag == "-" {
//			continue
//		}
//		fieldTag = strings.Replace(fieldTag, "column:", "", 1)
//		w.field2idx[fieldTag] = i
//	}
//	return
//}
//
//func (w *CopyEntity2MapWrapper) ConvertEntityToMap(data interface{}, fields []string) (output map[string]interface{}, skipfields []string, err error) {
//
//	output = map[string]interface{}{}
//
//	var val reflect.Value
//	if reflect.TypeOf(data).Kind() == reflect.Ptr {
//		val = reflect.ValueOf(data).Elem()
//	} else {
//		val = reflect.ValueOf(data)
//	}
//
//	if val.Type() != w.srcType {
//		err = fmt.Errorf("illegal data type||want=(%v) get=(%v)", w.srcType, val.Type())
//		return
//	}
//
//	for _, field := range fields {
//		if idx, exists := w.field2idx[field]; exists {
//			output[field] = val.Field(idx).Interface()
//		} else {
//			skipfields = append(skipfields, field)
//		}
//	}
//	return
//}
