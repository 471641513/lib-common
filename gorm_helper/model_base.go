package gorm_helper

import (
	"database/sql"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/opay-org/lib-common/utils"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/opay-org/lib-common/iowrapper/model"
	"github.com/opay-org/lib-common/local_context"
	"github.com/opay-org/lib-common/xlog"
	"github.com/smallnest/weighted"
)

const (
	DateActionType_Create = 1
	DateActionType_Update = 2
)

var commonReservedField = map[string]interface{}{
	"id": nil,
	"ID": nil,
	"Id": nil,
}

// for future extension
type ModelConfig struct {
	Prefix        string
	CntCacheSec   int
	ReservedField []string
}
type GormModelBase struct {
	db             *gorm.DB
	slaveDbs       []*gorm.DB
	cache          *model.BaseCacheModel
	w              *weighted.SW
	conf           *ModelConfig
	reservedFields map[string]interface{}
}

func NewGormModelBase(conf *ModelConfig, redisCli *redis.Client, dbConf model.DbConfig, slaveDbConf ...model.DbConfig) (m *GormModelBase, err error) {
	if conf == nil {
		conf = &ModelConfig{}
	}
	m = &GormModelBase{
		w:              &weighted.SW{},
		conf:           conf,
		reservedFields: map[string]interface{}{},
	}

	for field, _ := range commonReservedField {
		m.reservedFields[field] = nil
	}

	for _, field := range conf.ReservedField {
		m.reservedFields[field] = nil
	}

	m.cache = model.NewBaseCacheModel(redisCli, time.Minute*5)

	m.db, err = model.NewGorm(dbConf)
	if err != nil {
		xlog.Error("failed to NewOpayProductModel db||err=%v", err)
		return
	}
	idx := 0
	for _, _dbConf := range slaveDbConf {
		db, _err := model.NewGorm(_dbConf)
		if _err != nil {
			xlog.Warn("failed to init slave||skip||conf=%v||err=%v", _dbConf, err)
		}
		m.slaveDbs = append(m.slaveDbs, db)
		m.w.Add(idx, 1)
		idx += 1
	}
	return
}

func (m *GormModelBase) DB() *gorm.DB {
	return m.db
}

func (m *GormModelBase) SlaveDB() *gorm.DB {
	if len(m.slaveDbs) > 0 {
		idx := m.w.Next().(int)
		return m.slaveDbs[idx]
	}
	return m.db
}

func (m *GormModelBase) Cache() *model.BaseCacheModel {
	return m.cache
}

func (m *GormModelBase) ProcWriteAction(ctx *local_context.LocalContext, action ...DataWriteAction) (rslt *gorm.DB, err error) {
	return m.ProcWriteActions(ctx, action)
}

// process write action
func (m *GormModelBase) ProcWriteActions(ctx *local_context.LocalContext, actions []DataWriteAction) (rslt *gorm.DB, err error) {
	if len(actions) == 0 {
		err = fmt.Errorf("action is empty")
		return
	}

	defer func() {
		if rslt != nil && err == nil {
			if rslt.Error != nil {
				err = fmt.Errorf("db err=%v", rslt.Error)
			}
		}
		if err != nil {
			xlog.Error("logid=%v||failed to proc write action ||err=%v", ctx.LogId(), err)
		}
	}()

	if len(actions) == 1 {
		defer func() {
			if e := recover(); e != nil {
				xlog.Fatal("logid=%v||catch panic | %s\n%s", ctx.LogId(), e, debug.Stack())
				err = fmt.Errorf("panic||err=%v", e)
			}
			if err != nil {
				xlog.Error("logid=%v||MYSQL FAILED||err=%v", ctx.LogId(), err)
			}
		}()
		// process single
		action := actions[0]
		rslt, err = m.process(ctx, m.db, action)
		idx := 0
		if err != nil {
			xlog.Error("logid=%v||transaction failed at=%v||action=%+v||err=%v", ctx.LogId(), idx, action, err)
			return
		}
		if rslt == nil {
			err = fmt.Errorf("rslt nil")
			xlog.Error("logid=%v||transaction failed at=%v||action=%+v||err=%v", ctx.LogId(), idx, action, err)
			return
		}
		if rslt.Error != nil {
			err = fmt.Errorf("proc failed||err=%v", rslt.Error)
			xlog.Error("logid=%v||transaction failed at=%v||action=%+v||err=%v", ctx.LogId(), idx, action, err)
			return
		}
		if rslt.RowsAffected == 0 {
			xlog.Warn("logid=%v||no row affected||idx=%v||action=%+v", ctx.LogId(), idx, action)
			if !action.WriteActionBase().SkipEnsureRowAffected {
				err = fmt.Errorf("no row affected||where=%v||args=%v", action.WriteActionBase().Wheres, action.WriteActionBase().Args)
				return
			}
		}
	} else {
		tx := m.db.Begin()
		defer func() {
			if e := recover(); e != nil {
				xlog.Fatal("logid=%v||catch panic | %s\n%s", ctx.LogId(), e, debug.Stack())
				err = fmt.Errorf("panic||err=%v", e)
			}

			if err != nil {
				if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
					xlog.Error("logid=%v||ROLLBACK_FAILED||err=%v", ctx.LogId(), rollbackErr)
				} else {
					xlog.Error("logid=%v||MYSQL FAILED||err=%v", ctx.LogId(), err)
				}
			}
		}()
		// do transaction
		for idx, action := range actions {
			rslt, err = m.process(ctx, tx, action)
			if err != nil {
				xlog.Error("logid=%v||transaction failed at=%v||action=%+v||err=%v", ctx.LogId(), idx, action, err)
				return
			}
			if rslt == nil {
				err = fmt.Errorf("rslt nil")
				xlog.Error("logid=%v||transaction failed at=%v||action=%+v||err=%v", ctx.LogId(), idx, action, err)
				return
			}
			if rslt.Error != nil {
				err = fmt.Errorf("proc failed||err=%v", rslt.Error)
				xlog.Error("logid=%v||transaction failed at=%v||action=%+v||err=%v", ctx.LogId(), idx, action, err)
				return
			}
			if rslt.RowsAffected == 0 {
				xlog.Warn("logid=%v||no row affected||idx=%v||action=%+v", ctx.LogId(), idx, action)
				if !action.WriteActionBase().SkipEnsureRowAffected {
					err = fmt.Errorf("no row affected||where=%v||args=%v", action.WriteActionBase().Wheres, action.WriteActionBase().Args)
					return
				}
			}
		}
		tx.Commit()
	}
	return
}

func (m *GormModelBase) process(ctx *local_context.LocalContext, db *gorm.DB, a DataWriteAction) (dbRslt *gorm.DB, err error) {
	now := time.Now()
	if a.Entity() == nil {
		err = fmt.Errorf("entity is nil")
		return
	}

	switch a.WriteActionBase().Type {
	case DateActionType_Create:
		if e, ok := a.Entity().(entityWithUpdateTime); ok {
			e.SetUpdateTime(now.Unix())
		}

		dbRslt = db.Model(a.Entity())
		if onDuplicateIgnoreOpt := a.WriteActionBase().OnDuplicateOpt; onDuplicateIgnoreOpt != "" {
			dbRslt = dbRslt.Set("gorm:insert_option", fmt.Sprintf("ON DUPLICATE KEY %s", onDuplicateIgnoreOpt))
		}
		dbRslt = dbRslt.Create(a.Entity())
	case DateActionType_Update:
		tableName := ""
		updates := map[string]interface{}{}
		id := int64(0)
		var skip []string
		tableName = a.Entity().TableName()
		id = a.Entity().PrimaryId()
		updates, skip, err = a.Entity2MapWrapper().ConvertEntityToMap(a.Entity(), a.WriteActionBase().Fields)
		if err != nil {
			xlog.Error("logid=%v||faild to convert||err=%v", ctx.LogId(), err)
			return
		}

		if len(skip) > 0 {
			xlog.Warn("logid=%v||field skipped in convert||skip=%+v", ctx.LogId(), skip)
		}

		if !a.WriteActionBase().SkipEnsureRowAffected {
			if e, ok := a.Entity().(entityWithUpdateTime); ok {
				updates[e.UpdateTimeField()] = now
			}
		}
		// rm reserved field
		for f, _ := range m.reservedFields {
			if _, exists := updates[f]; exists {
				delete(updates, f)
			}
		}

		a.WriteActionBase().Wheres = append([]string{"id=?"}, a.WriteActionBase().Wheres...)
		a.WriteActionBase().Args = append([]interface{}{id}, a.WriteActionBase().Args...)
		dbRslt = db.Table(tableName).
			Where(strings.Join(a.WriteActionBase().Wheres, " and "), a.WriteActionBase().Args...).
			Updates(updates)

	default:
		err = fmt.Errorf("illegal action type=%v", a.WriteActionBase().Type)
	}
	return
}

func (m *GormModelBase) ProcQuery(
	ctx *local_context.LocalContext,
	action *DataQueryAction,
	out interface{}) (rslt *gorm.DB, totalCnt int64, err error) {

	if !action.SkipCount {
		c, err := m.CountQuery(ctx, action)
		if err != nil {
			return nil, 0, err
		}
		if c != nil {
			totalCnt = c.Total
		}
	}

	if action.SkipData {
		rslt = m.SlaveDB()
		return
	}

	//do query
	exec, err := action.prepareQuery(ctx, m.SlaveDB())
	if err != nil {
		return
	}
	rslt = exec.Offset(action.Offset).Limit(action.Limit)
	if action.SortByField != "" {
		rslt = rslt.Order(action.SortByField)
	}
	rslt = rslt.Find(out)
	err = rslt.Error
	return
}
func (m *GormModelBase) CountQuery(
	ctx *local_context.LocalContext,
	action *DataQueryAction) (c *DataQueryCache, err error) {
	c = &DataQueryCache{}
	key, err := action.cacheKey(totalCntCacheKey)

	xlog.Debug("logid=%v||action=%+v||key=%v||err=%v",
		ctx.LogId(), utils.MustString(action), key, err)
	if err != nil {
		return
	}

	c = &DataQueryCache{}
	if m.conf.CntCacheSec > 0 {
		exists, err := m.cache.Get(key, c)
		if err != nil {
			xlog.Error("logid=%v||failed to get count cache||err=%v", ctx.LogId(), err)
		}
		if exists {
			return c, err
		}
	}

	// count total
	exec, err := action.prepareQuery(ctx, m.SlaveDB())
	if err != nil {
		return
	}
	exec = exec.Count(&c.Total)
	if err := exec.Error; err != nil {
		xlog.Error("logid=%v||failed to count total||err=%v", ctx.LogId(), err)
		return c, err
	}
	if m.conf.CntCacheSec > 0 {
		err = m.cache.SetEx(key, c, time.Duration(m.conf.CntCacheSec)*time.Second)
		if err != nil {
			xlog.Error("logid=%v||failed to save cache||err=%v", ctx.LogId(), err)
		}
	}

	return
}

func (m *GormModelBase) ProcAggregation(
	ctx *local_context.LocalContext,
	action *DataAggregationAction) (output [][]ValAny, err error) {

	rslt := m.SlaveDB().Table(action.TableName).
		Select(strings.Join(action.Fields, ","))
	if action.Join != "" {
		rslt = rslt.Joins(action.Join)
	}
	rslt = rslt.Where(strings.Join(action.Wheres, " and "), action.Args...).
		Group(action.GroupBy)

	rows, err := rslt.Rows()
	if err != nil {
		xlog.Error("logid=%v||failed to do aggregation||err=%v", ctx.LogId(), err)
		return
	}

	fieldSize := len(action.Fields)

	for rows.Next() {
		itemPtr := make([]interface{}, fieldSize)
		for i := 0; i < fieldSize; i++ {
			var item sql.NullString
			itemPtr[i] = &item
		}
		err = rows.Scan(itemPtr...)
		tuple := make([]ValAny, fieldSize)
		for idx, v := range itemPtr {
			if v == nil {
				continue
			}
			if vBytes, ok := v.(*sql.NullString); ok {
				tuple[idx] = ValAny((*vBytes).String)
			}
		}
		if err != nil {
			xlog.Error("logid=%v||failed to scan rslt||err=%v", ctx.LogId(), err)
			return
		}

		output = append(output, tuple)
	}
	return
}
