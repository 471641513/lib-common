package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/opay-org/lib-common/xlog"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type DbConfig struct {
	Host     string `toml:"host"`
	Port     uint   `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Charset  string `toml:"charset"`
	Database string `toml:"database"`
	Debug    bool   `toml:"debug"`
	Loc      string `toml:"loc"`
}

func (c *DbConfig) GetDsn() string {
	if c.Loc == "" {
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=10s",
			c.User, c.Password, c.Host, c.Port, c.Database, c.Charset)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s&timeout=10s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.Charset, c.Loc)
}

type Logger struct {
}

func (l *Logger) Print(values ...interface{}) {
	if len(values) > 1 {
		source := values[1].(string)

		if dirs := strings.Split(source, "/"); len(dirs) >= 3 {
			source = strings.Join(dirs[len(dirs)-3:], "/")
		}

		if values[0] == "sql" {
			if len(values) > 5 {
				sql := gorm.LogFormatter(values...)[3]
				execTime := float64(values[2].(time.Duration).Nanoseconds()/1e4) / 100.0
				rows := values[5].(int64)
				xlog.Debug("query: <%s> | %.2fms | %d rows | %s", source, execTime, rows, sql)
			}
		} else {
			xlog.Debug("%v, %v", source, values[2:])
		}
	}
}

type LoggerWithTrace struct {
	TraceId string
}

func (l *LoggerWithTrace) Print(values ...interface{}) {
	if len(values) > 1 {
		source := values[1].(string)

		if dirs := strings.Split(source, "/"); len(dirs) >= 3 {
			source = strings.Join(dirs[len(dirs)-3:], "/")
		}

		if values[0] == "sql" {
			if len(values) > 5 {
				sql := gorm.LogFormatter(values...)[3]
				execTime := float64(values[2].(time.Duration).Nanoseconds()/1e4) / 100.0
				rows := values[5].(int64)
				xlog.Debug("%v | query: <%s> | %.2fms | %d rows | %s", l.TraceId, source, execTime, rows, sql)
			}
		} else {
			xlog.Debug("%v | %v, %v", l.TraceId, source, values[2:])
		}
	}
}

func NewGorm(config DbConfig) (orm *gorm.DB, err error) {
	dns := config.GetDsn()
	xlog.Info("dns=%v", dns)
	orm, err = gorm.Open("mysql", dns)
	if err != nil {
		return
	}
	if config.Debug {
		orm.LogMode(true)
	}
	orm.SetLogger(&Logger{})
	return
}
