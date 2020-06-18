package gorm_helper

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/opay-org/lib-common/local_context"
	"github.com/opay-org/lib-common/utils"
	"github.com/opay-org/lib-common/utils/obj_utils"
)

const maxLimit = 500
const totalCntCacheKey = "db:cnt"

type ValAny string

type Entity interface {
	TableName() string
	PrimaryId() int64
	Entity2MapWrapper() *obj_utils.CopyEntity2MapWrapper
}
type entityWithUpdateTime interface {
	UpdateTimeField() string
	SetUpdateTime(ts int64)
}

type DataWriteAction interface {
	Entity() Entity
	WriteActionBase() *DataWriteActionBase
}

type DataWriteActionBase struct {
	Type                  int
	Fields                []string      //更新字段
	Wheres                []string      // 更新condition
	Args                  []interface{} // where的args
	SkipEnsureRowAffected bool          // 强制全成功
}

func (b *DataWriteActionBase) WriteActionBase() *DataWriteActionBase {
	return b
}

type DataQueryAction struct {
	TableName   string
	Ids         []int64
	Fields      []string
	Wheres      []string
	Args        []interface{}
	Limit       int64
	Offset      int64
	SortByField string
	GroupBy     string
	Join        string

	SkipCount bool
	SkipData  bool

	where  []string
	args   []interface{}
	inited bool
}

type DataAggregationAction struct {
	TableName string
	Join      string
	Fields    []string
	Wheres    []string
	GroupBy   string
	Args      []interface{}
}

func (d *DataQueryAction) init() (err error) {
	if d.inited {
		return
	}
	where := d.Wheres
	args := d.Args
	if len(d.Ids) > 0 {
		where = append([]string{
			fmt.Sprintf(
				"id in (%s)",
				utils.IntListJoins(d.Ids, ","),
			),
		}, where...)
	}

	if len(d.Fields) > 0 {
		if d.GroupBy == "" {
			d.Fields = append(d.Fields, "id")
		}
	}
	if d.Join != "" {
		if len(d.Fields) == 0 {
			err = fmt.Errorf("illegal field length when join is given")
			return
		} else {
			for idx, field := range d.Fields {
				if !strings.Contains(field, " ") {
					d.Fields[idx] = fmt.Sprintf("%s.%s %s", d.TableName, field, field)
				}
			}
		}
	}
	d.where, d.args = where, args
	d.inited = true
	return
}

type DataQueryCache struct {
	Total int64 `json:"t"`
}

func (d *DataQueryAction) cacheKey(prefix string) (string, error) {
	err := d.init()
	if err != nil {
		return "", err
	}
	where := strings.Join(d.where, "")
	where = strings.ReplaceAll(where, "?", "%v")
	where = fmt.Sprintf(where, d.args...)
	where = strings.ReplaceAll(where, " ", "")
	if len(where) > 40 {
		where = utils.MD5Sum(where)
	}
	key := utils.GetKey(prefix+d.TableName, where)
	return key, nil
}

// TODO to simplify
func (d *DataQueryAction) prepareQuery(ctx *local_context.LocalContext, db *gorm.DB) (dbPrepare *gorm.DB, err error) {
	err = d.init()
	if err != nil {
		return
	}
	dbPrepare = db.Table(d.TableName)
	if len(d.Fields) > 0 {
		dbPrepare = dbPrepare.Select(d.Fields)
	}
	if d.Join != "" {
		dbPrepare = dbPrepare.Joins(d.Join)
	}

	if len(d.where) > 0 {
		dbPrepare = dbPrepare.Where(strings.Join(d.where, " and "), d.args...)
	}

	if d.GroupBy != "" {
		dbPrepare = dbPrepare.Group(d.GroupBy)
	}

	if d.Limit == 0 || d.Limit > maxLimit {
		d.Limit = maxLimit
	}
	return
}
