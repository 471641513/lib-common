package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/opay-org/lib-common/xlog"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func MD5Sum(in string) (out string) {
	ctx := md5.New()
	ctx.Write([]byte(in))
	return hex.EncodeToString(ctx.Sum(nil))
}

var timezoneLagos = time.FixedZone("UTC", 3600)

func GetDayStartTs(ts int64) (startTs int64) {
	t := time.Unix(ts, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, timezoneLagos).Unix()
}

func GetDayStr(ts int64) string {
	t := time.Unix(ts, 0)
	return t.Format("20060102")
}

func ConvertMapInt32(data interface{}) (mdata map[string]int32, err error) {
	switch t := data.(type) {
	case map[string]int:
		mdata = data.(map[string]int32)
	case map[string]int64:
		mrdata := data.(map[string]int64)
		if mrdata == nil {
			return
		}
		mdata = map[string]int32{}
		for k, v := range mrdata {
			mdata[k] = int32(v)
		}
	case map[string]int32:
		mrdata := data.(map[string]int64)
		if mrdata == nil {
			return
		}
		mdata = map[string]int32{}
		for k, v := range mrdata {
			mdata[k] = int32(v)
		}
	case map[string]interface{}:
		mrdata := data.(map[string]interface{})
		if mrdata == nil {
			return
		}
		mdata = map[string]int32{}
		for k, v := range mrdata {
			mdata[k] = int32(MustInt(v))
		}
	default:
		xlog.Error("illegal data type||data=%+v||data.type=%v", data, t)
		err = fmt.Errorf("illegal data type||data=%+v||data.type=%v", data, t)
	}
	return
}

func MustString(data interface{}) string {
	str, err := json.MarshalToString(data)
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return str
}
func MustStringIndent(data interface{}) string {
	bytes, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(bytes)
}

func MustInt(data interface{}) (i int) {
	var err error
	switch t := data.(type) {
	case float64:
		i = int(data.(float64))
	case float32:
		i = int(data.(float32))
	case int:
		i = data.(int)
	case int64:
		i = int(data.(int64))
	case int32:
		i = int(data.(int32))
	case string:
		i, err = strconv.Atoi(data.(string))
		if err != nil {
			xlog.Error("cant convert data to int||data=%+v", data)
			i = -1
		}
	default:
		xlog.Error("illegal data type||data=%+v||data.type=%v", data, t)
		i = -1
	}
	return
}

func MustBytes(data interface{}) (bytes []byte) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return []byte(fmt.Sprintf("%v", data))
	}
	return bytes
}

const (
	epsilon = 0.00001
)

func EarthDistance(lat1, lng1, lat2, lng2 float64) float64 {
	radius := float64(6371000) // 6378137
	rad := math.Pi / 180.0

	lat1 = lat1 * rad
	lng1 = lng1 * rad
	lat2 = lat2 * rad
	lng2 = lng2 * rad

	theta := lng2 - lng1

	if math.Abs(theta) < epsilon && math.Abs(lat2-lat1) < epsilon {
		return 0
	}

	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))

	return dist * radius
}

func CalTimecost(startTime time.Time) float64 {
	return float64(int(time.Now().Sub(startTime).Seconds() * 1000))
}

func convertStruct(a interface{}, b interface{}) error {
	data, err := json.Marshal(a)

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, b)

	if err != nil {
		return err
	}

	return nil
}

func GetKey(prefix string, items ...interface{}) string {
	format := prefix + strings.Repeat(":%v", len(items))
	return fmt.Sprintf(format, items...)
}

func ConvertStruct(a interface{}, b interface{}) error {
	err := convertStruct(a, b)

	if err != nil {
		xlog.Error("convert data failed | data: %s | error: %s", a, err)
	}

	return err
}

func ConvertStructs(items ...fmt.Stringer) (err error) {
	for i := 0; i < len(items)-1; i += 2 {
		if err := ConvertStruct(items[i], items[i+1]); err != nil {
			return err
		}
	}

	return
}

func CatchPanic() interface{} {
	if err := recover(); err != nil {
		xlog.Fatal("catch panic | %s\n%s", err, debug.Stack())
		return err
	}
	return nil
}

func FormatDate(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%4d%02d%02d", y, m, d)
}

func GetHourOneHot() []float32 {
	res := make([]float32, 24)
	res[time.Now().Hour()] = 1
	return res
}

func GetCurrentDayStartTime() (timestamp int64, err error) {
	timeLayout := "2006-01-02"
	tm := time.Unix(time.Now().Unix(), 0)
	dateStr := tm.Format(timeLayout)
	theTime, err := time.Parse(timeLayout, dateStr)
	if err != nil {
		return
	}
	timestamp = theTime.Unix()
	return
}

func Bool2Int(val bool) (res int) {
	if val == true {
		res = 1
	}
	return res
}

func UserRegTransfer(curTimestamp int64, userRegTime int64) (res int) {
	gapVal := curTimestamp - userRegTime
	if gapVal <= 0 {

	} else if 0 < gapVal && gapVal <= 84000*7 {
		res = 1
	} else if gapVal <= 30*86400 {
		res = 2
	} else {
		res = 3
	}
	return res
}

func Str2int64(s string) int64 {
	var res int64
	if s != "" {
		x, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			res = x
		}
	}
	return res
}
