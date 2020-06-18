package gorm_helper

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/opay-org/lib-common/utils/obj_utils"

	"github.com/jinzhu/gorm"
	"github.com/opay-org/lib-common/iowrapper/model"
	"github.com/opay-org/lib-common/iowrapper/redis_wrapper"
	"github.com/opay-org/lib-common/local_context"
	"github.com/opay-org/lib-common/utils"
	"github.com/opay-org/lib-common/xlog"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	xlog.SetupLogDefault()
	if err := setUp(); err != nil {
		xlog.Close()
		xlog.Error("failed to do test setup")
		os.Exit(255)
	}
	// setup code...
	code := m.Run()
	// teardown code...
	xlog.Close()
	os.Exit(code)
}

type testDataModel struct {
	*GormModelBase
}

var defaultM *testDataModel

func setUp() (err error) {
	redisCli, err := redis_wrapper.NewRedisClient(&redis_wrapper.RedisConfig{
		Addrs: []string{"127.0.0.1:6379"},
	})
	if err != nil {
		return
	}

	defaultM = &testDataModel{}
	defaultM.GormModelBase, err = NewGormModelBase(
		nil,
		redisCli,
		model.DbConfig{
			Host:     "127.0.0.1",
			Port:     3306,
			Password: "123456",
			User:     "root",
			Database: "test",
			Debug:    true,
			Charset:  "utf8",
		})

	if err != nil {
		return
	}
	defaultM.DB().Delete(&testEntity{}, "id<10000")
	return
}

//CREATE TABLE `data_test` (
//`id` bigint NOT NULL DEFAULT '0',
//`ori_addr` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
//`ori_lat` decimal(10,6) DEFAULT NULL,
//`ori_lng` decimal(10,6) DEFAULT NULL,
//`update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
//`status` int DEFAULT NULL,
//PRIMARY KEY (`id`)
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

type testEntity struct {
	ID         int64     `gorm:"column:id" json:"id"`
	OriAddr    string    `gorm:"column:ori_addr"  json:"ori_addr"` // "ori1",
	OriLat     float64   `gorm:"column:ori_lat" json:"ori_lat"`    //  1,
	OriLng     float64   `gorm:"column:ori_lng" json:"ori_lng"`    //  2,
	Status     int32     `gorm:"column:status" json:"status"`
	UpdateTime time.Time `gorm:"column:update_time" json:"update_time"`
}

func (e *testEntity) AfterFind() (err error) {
	xlog.Fatal("e=%+v", e)
	return
}
func (e *testEntity) TableName() string {
	return "data_test"
}

func (e *testEntity) PrimaryId() int64 {
	return e.ID
}
func (e *testEntity) PrimarySeted() bool {
	return e.ID > 0
}

func (e *testEntity) UpdateTimeField() string {
	return "update_time"
}

func (e *testEntity) SetUpdateTime(ts int64) {
	e.UpdateTime = time.Unix(ts, 0)
}

func (e *testEntity) Entity() Entity {
	return e
}

var convertWrapper *obj_utils.CopyEntity2MapWrapper

func init() {
	var err error
	convertWrapper, err = obj_utils.CompileCopyEntity2MapWrapper(testEntity{})
	if err != nil {
		xlog.Fatal("failed to build convertWrapper||err=%v", err)
	} else {
		xlog.Info("wrapper=%+v", convertWrapper)
	}
}

func (e *testDataAction) Entity2MapWrapper() *obj_utils.CopyEntity2MapWrapper {
	return convertWrapper
}

type testDataAction struct {
	*testEntity
	*DataWriteActionBase
}

type statEntity struct {
	Cnt int64 `json:"cnt"`
}

func TestTestDataModel_ProcQuery(t *testing.T) {

	stat := []*statEntity{}
	action := &DataQueryAction{
		GroupBy:   "ori_addr",
		TableName: "data_test",
		Fields: []string{
			"count(1) as cnt",
		},
		SortByField: "",
		Limit:       10,
	}
	ctx := local_context.NewLocalContext()
	_, totalCnt, err := defaultM.ProcQuery(ctx, action, &stat)
	xlog.Info("err=%+v||totalCnt=%v||perf=%+v", err, totalCnt, utils.MustString(stat))
}

func TestUpdateTimestampTimezone(t *testing.T) {
	ctx := local_context.NewLocalContext()
	type args struct {
		actions []DataWriteAction
		assert  func(t *testing.T, rslt *gorm.DB, err error)
	}

	nowTs := time.Now()
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create single order",
			args: args{
				actions: []DataWriteAction{
					&testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Create,
						},
						testEntity: &testEntity{
							ID:      233,
							OriAddr: "ori1",
							OriLat:  1,
							OriLng:  2,
						},
					}},
				assert: func(t *testing.T, rslt *gorm.DB, err error) {
					assert.Nil(t, err)
					assert.Nil(t, rslt.Error)
					o := &testEntity{}
					defaultM.DB().Where("id=233").Find(o)
					assert.Equal(t, o.OriAddr, "ori1")
					xlog.Info("e=%+v", utils.MustString(o))
					//assert.Equal(t, nowTs.Unix(), o.UpdateTime.Unix())

					o2 := []*testEntity{}
					query := &DataQueryAction{
						TableName: o.TableName(),
						Wheres: []string{
							fmt.Sprintf("unix_timestamp(update_time)<=%v", nowTs.Unix()+1),
						},
					}
					rrslt, _, err := defaultM.ProcQuery(ctx, query, &o2)
					assert.Nil(t, err)
					assert.Nil(t, rrslt.Error)
					xlog.Info("o2=%+v", o2)

					assert.Equal(t, "ori1", o2[0].OriAddr)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRslt, err := defaultM.ProcWriteActions(ctx, tt.args.actions)
			tt.args.assert(t, gotRslt, err)

		})
	}

}

func TestTestDataModel_ProcWriteActions(t *testing.T) {
	ctx := local_context.NewLocalContext()
	type args struct {
		actions []DataWriteAction
		assert  func(t *testing.T, rslt *gorm.DB, err error)
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "create single order",
			args: args{
				actions: []DataWriteAction{
					&testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Create,
						},
						testEntity: &testEntity{
							ID:      1,
							OriAddr: "ori1",
							OriLat:  1,
							OriLng:  2,
						},
					}},
				assert: func(t *testing.T, rslt *gorm.DB, err error) {
					assert.Nil(t, err)
					assert.Nil(t, rslt.Error)
					o := &testEntity{}
					defaultM.DB().Where("id=1").Find(o)
					assert.Equal(t, o.OriAddr, "ori1")
					// query by method
					o2 := []*testEntity{}
					query := &DataQueryAction{
						TableName: o.TableName(),
						Ids:       []int64{1},
					}
					rrslt, _, err := defaultM.ProcQuery(ctx, query, &o2)
					assert.Nil(t, err)
					assert.Nil(t, rrslt.Error)
					xlog.Info("o2=%+v", o2)

					assert.Equal(t, o2[0].OriAddr, "ori1")
				},
			},
		}, {
			name: "create single order fail",
			args: args{
				actions: []DataWriteAction{
					&testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type:           DateActionType_Create,
							OnDuplicateOpt: "update id=id",
						},
						testEntity: &testEntity{
							ID:      1,
							OriAddr: "ori1",
							OriLat:  1,
							OriLng:  2,
						},
					}},
				assert: func(t *testing.T, rslt *gorm.DB, err error) {
					assert.NotNil(t, err)
				},
			},
		},
		{
			name: "create single order on duplicate update",
			args: args{
				actions: []DataWriteAction{
					&testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type:                  DateActionType_Create,
							OnDuplicateOpt:        "update id=id",
							SkipEnsureRowAffected: true,
						},
						testEntity: &testEntity{
							ID:      1,
							OriAddr: "ori1",
							OriLat:  1,
							OriLng:  2,
						},
					}},
				assert: func(t *testing.T, rslt *gorm.DB, err error) {
					assert.Nil(t, err)
					assert.Nil(t, rslt.Error)
				},
			},
		},
		{
			name: "update single order",
			args: args{
				actions: []DataWriteAction{
					&testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Update,
							Fields: []string{
								"ori_addr",
								"ori_lng",
							},
						},
						testEntity: &testEntity{
							ID:      1,
							OriAddr: "ori2",
							OriLat:  10,
							OriLng:  20,
						},
					},
				},
				assert: func(t *testing.T, rslt *gorm.DB, err error) {
					assert.Nil(t, err)
					assert.Nil(t, rslt.Error)
					o := &testEntity{}
					defaultM.DB().Where("id=1").Find(o)
					assert.Equal(t, o.OriAddr, "ori2")
					assert.Equal(t, o.OriLat, float64(1))
					assert.Equal(t, o.OriLng, float64(20))
				},
			},
		},
		{
			name: "test roll back by mysql failed",
			args: args{
				actions: []DataWriteAction{
					&testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Create,
						},
						testEntity: &testEntity{
							ID:      2,
							OriAddr: "rollback_1",
							OriLat:  1,
							OriLng:  2,
						},
					}, &testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Create,
						},
						testEntity: &testEntity{
							ID:      2,
							OriAddr: "rollback_1",
							OriLat:  1,
							OriLng:  2,
						},
					}},
				assert: func(t *testing.T, rslt *gorm.DB, err error) {
					assert.True(t, err != nil || rslt.Error != nil)
					o := &testEntity{}
					defaultM.DB().Where("id=2").Find(o)
					xlog.Info("o=%+v", o)
					assert.True(t, !o.PrimarySeted())
				},
			},
		},
		{
			name: "test roll back - by no update but match hit",
			args: args{
				actions: []DataWriteAction{
					&testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Create,
						},
						testEntity: &testEntity{
							ID:      2,
							OriAddr: "rollback_1",
							OriLat:  1,
							OriLng:  2,
							Status:  1,
						},
					}, &testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Update,
							Fields: []string{
								"status",
								"ori_lng",
							},
							Wheres: []string{
								"status in (1,2)",
							},
						},
						testEntity: &testEntity{
							ID:     2,
							OriLat: 10,
							OriLng: 20,
							Status: 2,
						},
					}},
				assert: func(t *testing.T, rslt *gorm.DB, err error) {
					assert.Nil(t, err)
					assert.Nil(t, rslt.Error)
					o := &testEntity{}
					defaultM.DB().Where("id=2").Find(o)
					xlog.Info("o=%+v", o)
					assert.True(t, o.PrimarySeted())
					assert.Equal(t, int32(2), o.Status)
					assert.Equal(t, float64(1), o.OriLat)
					assert.Equal(t, float64(20), o.OriLng)
					/*
						assert.True(t, err != nil || rslt.Error != nil)

						o := &testEntity{}
						defaultM.DB().Where("id=2").Find(o)
						xlog.Info("o=%+v", o)
						assert.True(t, !o.PrimarySeted())
					*/
				},
			},
		},
		{
			name: "test roll back - by no update and no match hit - update and update",
			args: args{
				actions: []DataWriteAction{
					&testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Update,
							Fields: []string{
								"status",
								"ori_addr",
							},
						},
						testEntity: &testEntity{
							ID:      2,
							OriAddr: "rollback_4",
							Status:  100,
						},
					}, &testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Update,
							Fields: []string{
								"status",
								"ori_addr",
							},
							Wheres: []string{
								"status in (3,2)",
							},
						},
						testEntity: &testEntity{
							ID:      2,
							OriAddr: "rollback_4",
							Status:  100,
						},
					}},
				assert: func(t *testing.T, rslt *gorm.DB, err error) {
					assert.True(t, err != nil || rslt.Error != nil)
					o := &testEntity{}
					defaultM.DB().Where("id=2").Find(o)
					xlog.Info("o=%+v", o)
					assert.Equal(t, o.Status, int32(2))
					assert.Equal(t, o.OriAddr, "rollback_1")
				},
			},
		},
		{
			name: "test roll back - by no update and no match hit - create and update",
			args: args{
				actions: []DataWriteAction{
					&testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Create,
						},
						testEntity: &testEntity{
							ID:      3,
							OriAddr: "rollback_3",
							OriLat:  1,
							OriLng:  2,
							Status:  1,
						},
					}, &testDataAction{
						DataWriteActionBase: &DataWriteActionBase{
							Type: DateActionType_Update,
							Fields: []string{
								"status",
								"ori_lng",
							},
							Wheres: []string{
								"status in (3,2)",
							},
						},
						testEntity: &testEntity{
							ID:     3,
							OriLat: 10,
							OriLng: 20,
							Status: 2,
						},
					}},
				assert: func(t *testing.T, rslt *gorm.DB, err error) {
					assert.True(t, err != nil || rslt.Error != nil)
					o := &testEntity{}
					defaultM.DB().Where("id=3").Find(o)
					xlog.Info("o=%+v", o)
					assert.True(t, !o.PrimarySeted())
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRslt, err := defaultM.ProcWriteActions(ctx, tt.args.actions)
			tt.args.assert(t, gotRslt, err)

		})
	}
}
