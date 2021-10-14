package obj_utils

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/xutils/lib-common/xlog"
)

const (
	numericalInt = iota
	numericalUInt
	numericalFloat
	numericalTime
)

var numericalFields = map[reflect.Kind]int{
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
}

func getIntValidRange(t reflect.Type) (maxValid, minValid int64) {
	switch t.Kind() {
	case reflect.Int8:
		maxValid = math.MaxInt8
		minValid = -math.MaxInt8
	case reflect.Int16:
		maxValid = math.MaxInt16
		minValid = -math.MaxInt16
	case reflect.Int32:
		maxValid = math.MaxInt32
		minValid = -math.MaxInt32
	case reflect.Int64:
		maxValid = math.MaxInt64
		minValid = -math.MaxInt64
	case reflect.Int:
		maxValid = math.MaxInt64
		minValid = -math.MaxInt64
	default:
		panic("invalid visit")
	}
	return
}

func getUintValidRange(t reflect.Type) (maxValid, minValid uint64) {
	switch t.Kind() {
	case reflect.Uint8:
		maxValid = math.MaxUint8
	case reflect.Uint16:
		maxValid = math.MaxUint16
	case reflect.Uint32:
		maxValid = math.MaxUint32
	case reflect.Uint64:
		maxValid = math.MaxUint64
	case reflect.Uint:
		maxValid = math.MaxUint64
	default:
		panic("invalid visit")
	}
	return
}
func getFloatValidRange(t reflect.Type) (maxValid, minValid float64) {
	switch t.Kind() {
	case reflect.Float32:
		maxValid = math.MaxFloat32
		minValid = -math.MaxFloat32
	case reflect.Float64:
		maxValid = math.MaxFloat64
		minValid = -math.MaxFloat64
	default:
		panic("invalid visit")
	}
	return
}

func getValidFunc(srcType reflect.Type, destType reflect.Type, srcNt, destNt int) (f func(srcVal reflect.Value) (err error)) {

	switch srcNt {
	case numericalUInt:
	case numericalFloat:
	case numericalInt:
	case numericalTime:
	default:
	}

	switch destNt {
	case numericalUInt:
		switch srcNt {
		case numericalUInt:
			srcMax, _ := getUintValidRange(srcType)
			destMax, _ := getUintValidRange(destType)
			if destMax < srcMax {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Uint() > destMax {
						err = fmt.Errorf("src val exceed max valid val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Uint(), srcType, destType)
					}
					return
				}
			}

		case numericalFloat:
			srcMax, _ := getFloatValidRange(srcType)
			destMax, _ := getUintValidRange(destType)

			if destMax >= uint64(srcMax) {
				xlog.Warn("possible precision loss if src is negative||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Float() < float64(0) {
						err = fmt.Errorf("src val exceed m min val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Float(), srcType, destType)
						return
					}
					return
				}
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Float() < float64(0) {
						err = fmt.Errorf("src val exceed m min val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcType, srcVal.Float(), destType)
						return
					}
					if uint64(srcVal.Float()) > destMax {
						err = fmt.Errorf("src val exceed m min val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Float(), srcType, destType)
						return
					}
					return
				}
			}

		case numericalInt:
			srcMax, _ := getIntValidRange(srcType)
			destMax, _ := getUintValidRange(destType)

			if destMax >= uint64(srcMax) {
				xlog.Warn("possible precision loss if src is negative||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Int() < 0 {
						err = fmt.Errorf("src val exceed m min val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Int(), srcType, destType)
						return
					}
					return
				}
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Int() < 0 {
						err = fmt.Errorf("src val exceed m min val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Int(), srcType, destType)
						return
					}
					if uint64(srcVal.Int()) > destMax {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Int(), srcType, destType)
						return
					}
					return
				}
			}
		case numericalTime:
			srcMax := math.MaxInt64
			destMax, _ := getUintValidRange(destType)

			if destMax >= uint64(srcMax) {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {

					val := time.Time{}
					reflect.ValueOf(&val).Elem().Set(srcVal)
					if uint64(val.Unix()) > destMax {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, val.Unix(), srcType, destType)
						return
					}
					return
				}
			}
		default:
		}
	case numericalFloat:
		switch srcNt {
		case numericalUInt:
			srcMax, _ := getUintValidRange(srcType)
			destMax, _ := getFloatValidRange(destType)
			if uint64(destMax) >= srcMax {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Uint() > uint64(destMax) {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Uint(), srcType, destType)
						return
					}
					return
				}
			}
		case numericalFloat:
			srcMax, _ := getFloatValidRange(srcType)
			destMax, _ := getFloatValidRange(destType)
			if destMax >= srcMax {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Float() > destMax {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Float(), srcType, destType)
						return
					}
					return
				}
			}
		case numericalInt:
			srcMax, _ := getIntValidRange(srcType)
			destMax, _ := getFloatValidRange(destType)
			if destMax >= float64(srcMax) {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if float64(srcVal.Int()) > destMax {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Int(), srcType, destType)
						return
					}
					return
				}
			}
		case numericalTime:
			srcMax := math.MaxInt64
			destMax, _ := getFloatValidRange(destType)

			if destMax >= float64(srcMax) {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {

					val := time.Time{}
					reflect.ValueOf(&val).Elem().Set(srcVal)
					if float64(val.Unix()) > destMax {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, val.Unix(), srcType, destType)
						return
					}
					return
				}
			}

		default:
		}
	case numericalInt:
		switch srcNt {
		case numericalUInt:
			srcMax, _ := getUintValidRange(srcType)
			destMax, _ := getIntValidRange(destType)
			if uint64(destMax) >= srcMax {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Uint() > uint64(destMax) {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Uint(), srcType, destType)
						return
					}
					return
				}
			}
		case numericalFloat:
			srcMax, _ := getFloatValidRange(srcType)
			destMax, _ := getIntValidRange(destType)
			if float64(destMax) >= srcMax {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Float() > float64(destMax) {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Float(), srcType, destType)
						return
					}
					return
				}
			}
		case numericalInt:
			srcMax, _ := getIntValidRange(srcType)
			destMax, _ := getIntValidRange(destType)
			if destMax >= srcMax {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Int() > destMax {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Int(), srcType, destType)
						return
					}
					return
				}
			}
		case numericalTime:
			srcMax := math.MaxInt64
			destMax, _ := getIntValidRange(destType)
			if destMax >= int64(srcMax) {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {

					val := time.Time{}
					reflect.ValueOf(&val).Elem().Set(srcVal)
					if val.Unix() > destMax {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, val.Unix(), srcType, destType)
						return
					}
					return
				}
			}
		default:
		}
	case numericalTime:
		switch srcNt {
		case numericalUInt:
			srcMax, _ := getUintValidRange(srcType)
			destMax := math.MaxInt64
			if uint64(destMax) >= srcMax {
				// do nothing
			} else {
				xlog.Warn("possible precision loss if src value exceed dest||srcType=%+v||destType=%+v",
					srcType, destType)
				f = func(srcVal reflect.Value) (err error) {
					if srcVal.Uint() > uint64(destMax) {
						err = fmt.Errorf("src val exceed m max val for dest||max=%v||val=%v||srcType=%+v||destType=%+v",
							destMax, srcVal.Uint(), srcType, destType)
						return
					}
					return
				}
			}

		case numericalFloat:
			// do nothing
		case numericalInt:
			// do nothing
		case numericalTime:
			// do nothing
		default:
		}
	default:
		panic("illegal type")
	}

	return
}

func getConverFunc(srcType reflect.Type, destType reflect.Type, srcNt, destNt int) (convertFunc func(srcVal reflect.Value, destVal reflect.Value) (err error)) {
	switch destNt {
	case numericalTime:
		switch srcNt {
		case numericalInt:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				t := time.Unix(int64(srcVal.Int()), 0)
				destVal.Set(reflect.ValueOf(t))
				return
			}
		case numericalUInt:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				t := time.Unix(int64(srcVal.Uint()), 0)
				destVal.Set(reflect.ValueOf(t))
				return
			}
		case numericalFloat:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				t := time.Unix(int64(srcVal.Float()), 0)

				destVal.Set(reflect.ValueOf(t))
				return
			}
		case numericalTime:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				destVal.Set(srcVal)
				return
			}
		}
	case numericalInt:
		switch srcNt {
		case numericalInt:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {

				destVal.SetInt(int64(srcVal.Int()))
				return
			}
		case numericalUInt:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {

				destVal.SetInt(int64(srcVal.Uint()))
				return
			}
		case numericalFloat:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {

				destVal.SetInt(int64(srcVal.Float()))
				return
			}
		case numericalTime:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				val := time.Time{}
				reflect.ValueOf(&val).Elem().Set(srcVal)
				destVal.SetInt(val.Unix())
				return
			}
		}
	case numericalUInt:
		switch srcNt {
		case numericalInt:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {

				destVal.SetUint(uint64(srcVal.Int()))
				return
			}
		case numericalUInt:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				destVal.SetUint(uint64(srcVal.Uint()))
				return
			}
		case numericalFloat:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				destVal.SetUint(uint64(srcVal.Float()))
				return
			}
		case numericalTime:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				val := time.Time{}
				reflect.ValueOf(&val).Elem().Set(srcVal)
				destVal.SetUint(uint64(val.Unix()))
				return
			}
		}

	case numericalFloat:
		switch srcNt {
		case numericalInt:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				destVal.SetFloat(float64(srcVal.Int()))
				return
			}
		case numericalUInt:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {

				destVal.SetFloat(float64(srcVal.Uint()))
				return
			}
		case numericalFloat:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				destVal.SetFloat(float64(srcVal.Float()))
				return
			}
		case numericalTime:
			convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				val := time.Time{}
				reflect.ValueOf(&val).Elem().Set(srcVal)
				destVal.SetFloat(float64(val.Unix()))
				return
			}
		}

	default:
		panic("illegal type")
	}
	return
}

var reservedField = map[string]interface{}{
	"id_str": nil,
	"id":     nil,
}

var reservedNumericalStruct = map[reflect.Type]int{
	reflect.TypeOf(time.Time{}): numericalTime,
}

type copyMapping struct {
	name        string
	alias       string
	srcIdx      int
	srcSafeMax  uint64
	srcNtStr    string
	destNtStr   string
	validFunc   func(ori reflect.Value) error
	convertFunc func(dest reflect.Value, ori reflect.Value) error
}

type CopyFieldWrapper struct {
	srcType  reflect.Type
	destType reflect.Type

	oriName2idx  map[string]int
	destName2idx map[string]int
	mapping      map[int]copyMapping
}

func CompileCopyFieldWrapper(src interface{}, dest interface{}) (w *CopyFieldWrapper, err error) {
	if src == nil || dest == nil {
		err = fmt.Errorf("src or dest nil||src=%v||dest=%v", src, dest)
		return
	}

	if reflect.TypeOf(src).Kind() != reflect.Struct {
		xlog.Warn("reflect.TypeOf(src).Kind() =%v", reflect.TypeOf(src).Kind())
		err = fmt.Errorf("src not struct")
		return
	}

	if reflect.TypeOf(dest).Kind() != reflect.Struct {
		err = fmt.Errorf("dest not struct")
		return
	}
	return compileCopyFieldWrapper(reflect.TypeOf(src), reflect.TypeOf(dest))
}

func compileCopyFieldWrapper(srcType, destType reflect.Type) (w *CopyFieldWrapper, err error) {
	w = &CopyFieldWrapper{
		srcType:  srcType,
		destType: destType,
		mapping:  map[int]copyMapping{},
	}

	// generate mapping

	//srcv := oriSrc
	srct := w.srcType
	//destv := oriDest
	destt := w.destType

	// 1. generate json tag to idx mapping for src
	srcFieldJsonOrName2Idx := map[string]int{}
	for i := 0; i < srct.NumField(); i++ {
		field := srct.Field(i)

		if field.Type.Kind() == reflect.Struct ||
			field.Type.Kind() == reflect.Ptr {
			if _, exists := reservedNumericalStruct[field.Type]; !exists {
				continue
			}
		}

		jtag := getJtag(field)
		if jtag == "-" {
			continue
		}
		if jtag != "" {
			srcFieldJsonOrName2Idx[jtag] = i
		}
		srcFieldJsonOrName2Idx[field.Name] = i
	}

	w.oriName2idx = srcFieldJsonOrName2Idx

	destFieldJsonOrName2Idx := map[string]int{}
	for i := 0; i < destt.NumField(); i++ {
		field := destt.Field(i)

		if field.Type.Kind() == reflect.Struct ||
			field.Type.Kind() == reflect.Ptr {
			if _, exists := reservedNumericalStruct[field.Type]; !exists {
				continue
			}
		}

		jtag := getJtag(field)
		if jtag == "-" {
			continue
		}
		if jtag == "-" {
			continue
		}
		if jtag != "" {
			destFieldJsonOrName2Idx[jtag] = i
		}
		destFieldJsonOrName2Idx[field.Name] = i
	}

	w.destName2idx = destFieldJsonOrName2Idx

	// 2. loop dest obj
	for i := 0; i < destt.NumField(); i++ {
		field := destt.Field(i)

		if field.Type.Kind() == reflect.Struct ||
			field.Type.Kind() == reflect.Ptr {
			if _, exists := reservedNumericalStruct[field.Type]; !exists {
				continue
			}
		}

		m := copyMapping{}

		m.name = field.Name
		m.alias = getJtag(field)
		var exists bool
		m.srcIdx, exists = srcFieldJsonOrName2Idx[m.name]
		if !exists {
			m.srcIdx, exists = srcFieldJsonOrName2Idx[m.alias]
			if !exists {
				continue
			}
		}
		srcField := srct.Field(m.srcIdx)

		if srcField.Type == field.Type {
			m.convertFunc = func(destVal reflect.Value, srcVal reflect.Value) (err error) {
				destVal.Set(srcVal)
				return
			}
		} else {
			srcNt, exist := numericalFields[srcField.Type.Kind()]
			if !exist {
				if srcNt, exist = reservedNumericalStruct[srcField.Type]; !exist {
					continue
				}
			}
			destNt, exist := numericalFields[field.Type.Kind()]
			if !exist {
				if destNt, exist = reservedNumericalStruct[field.Type]; !exist {
					continue
				}
			}

			m.srcNtStr = srcField.Type.String()
			m.destNtStr = field.Type.String()

			// build valid func
			m.validFunc = getValidFunc(srcField.Type, field.Type, srcNt, destNt)
			m.convertFunc = getConverFunc(srcType, destType, srcNt, destNt)

		}
		if m.convertFunc == nil {
			xlog.Warn("convertFunc nil||m=%+v", m)
			continue
		}
		w.mapping[i] = m
	}
	return
}

func (w *CopyFieldWrapper) SkippedFields() (fields []string) {
	destt := w.destType
	for i := 0; i < destt.NumField(); i++ {
		field := destt.Field(i)
		if field.Type.Kind() == reflect.Struct ||
			field.Type.Kind() == reflect.Ptr {
			continue
		}
		jtag := getJtag(field)
		if jtag == "-" {
			continue
		}
		if _, exists := w.mapping[i]; !exists {
			fields = append(fields, field.Name)
		}
	}
	return
}

var aggreFieldExp *regexp.Regexp

func init() {
	aggreFieldExp, _ = regexp.Compile("^(.*\\s+)")
}

func (w *CopyFieldWrapper) CopyFieldValues(srcPtr interface{}, destPtr interface{}, projection ...string) (skipFields []string, err error) {
	if srcPtr == nil || destPtr == nil {
		err = fmt.Errorf("src or dest nil||src=%v||dest=%v", srcPtr, destPtr)
		return
	}
	if reflect.TypeOf(srcPtr).Kind() != reflect.Ptr {
		err = fmt.Errorf("src not ptr")
		return
	}

	if reflect.TypeOf(destPtr).Kind() != reflect.Ptr {
		err = fmt.Errorf("dest not ptr")
		return
	}

	if reflect.ValueOf(srcPtr).IsNil() {
		err = fmt.Errorf("srcPtr is nil")
		return
	}

	if reflect.ValueOf(destPtr).IsNil() {
		err = fmt.Errorf("destPtr is nil")
		return
	}

	oriSrc := reflect.ValueOf(srcPtr).Elem()
	if oriSrc.Type() != w.srcType {
		err = fmt.Errorf("illegal oriSrc type||want(%v) get(%v)", w.srcType, oriSrc.Type())
		return
	}
	oriDest := reflect.ValueOf(destPtr).Elem()
	if oriDest.Type() != w.destType {
		err = fmt.Errorf("illegal destSrc type||want(%v) get(%v)", w.destType, oriDest.Type())
		return
	}

	// generate projection
	withProj := len(projection) > 0

	// 1. do copy
	if withProj {
	reservFieldLoop:
		for field, _ := range reservedField {
			if idx, exists := w.destName2idx[field]; exists {
				if m, exists := w.mapping[idx]; exists && m.convertFunc != nil {
					// 3. copy value
					oriVal := oriSrc.Field(m.srcIdx)
					destVal := oriDest.Field(idx)
					if m.validFunc != nil {
						if err := m.validFunc(oriVal); err != nil {
							xlog.Warn("failed to valid value||field=%+v||m=%+v||err=%v", field, m, err)
							skipFields = append(skipFields, field)
							continue reservFieldLoop
						}
					}
					err := m.convertFunc(destVal, oriVal)
					if err != nil {
						xlog.Error("failed to copy value||field=%+v||m=%+v||err=%v", field, m, err)
						skipFields = append(skipFields, field)
					}
				}
			}
		}
	projLoop:
		for _, field := range projection {
			field = aggreFieldExp.ReplaceAllString(field, "")
			if idx, exists := w.destName2idx[field]; exists {
				if m, exists := w.mapping[idx]; exists && m.convertFunc != nil {
					// 3. copy value
					oriVal := oriSrc.Field(m.srcIdx)
					destVal := oriDest.Field(idx)
					if m.validFunc != nil {
						if err := m.validFunc(oriVal); err != nil {
							xlog.Warn("failed to valid value||field=%+v||err=%v", field, err)
							skipFields = append(skipFields, field)
							continue projLoop
						}
					}
					err := m.convertFunc(destVal, oriVal)
					if err != nil {
						xlog.Error("failed to copy value||field=%+v||err=%v", field, err)
						skipFields = append(skipFields, field)
						continue projLoop
					}
				} else {
					skipFields = append(skipFields, field)
				}
			} else {
				skipFields = append(skipFields, field)
			}
		}
	} else {
	fieldLoop:
		for i := 0; i < w.destType.NumField(); i++ {
			field := w.destType.Field(i)
			if m, exists := w.mapping[i]; exists && m.convertFunc != nil {
				// 3. copy value
				oriVal := oriSrc.Field(m.srcIdx)
				destVal := oriDest.Field(i)
				if m.validFunc != nil {
					if err := m.validFunc(oriVal); err != nil {
						xlog.Warn("failed to valid value||field=%+v||err=%v", field.Name, err)
						skipFields = append(skipFields, field.Name)
						continue fieldLoop
					}
				}
				err := m.convertFunc(destVal, oriVal)
				if err != nil {
					xlog.Error("failed to copy value||field=%+v||err=%v", field.Name, err)
					skipFields = append(skipFields, field.Name)
					continue fieldLoop
				}
			} else {
				skipFields = append(skipFields, field.Name)
			}
		}
	}
	return
}

func getJtag(field reflect.StructField) (tag string) {
	tag = field.Tag.Get("json")
	l := strings.SplitN(tag, ",", 2)
	if len(l) > 0 {
		tag = l[0]
	}
	return
}

func copyFieldValues(src interface{}, dest interface{}, projection ...string) (skipFields []string, err error) {

	// 0. build wrapper
	if src == nil || dest == nil {
		err = fmt.Errorf("src or dest nil||src=%v||dest=%v", src, dest)
		return
	}
	if reflect.TypeOf(src).Kind() != reflect.Ptr {
		err = fmt.Errorf("src not ptr")
		return
	}

	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		err = fmt.Errorf("dest not ptr")
		return
	}

	w, err := compileCopyFieldWrapper(reflect.ValueOf(src).Elem().Type(), reflect.ValueOf(dest).Elem().Type())
	if err != nil {
		return
	}
	skipFields, err = w.CopyFieldValues(src, dest, projection...)
	return
}
