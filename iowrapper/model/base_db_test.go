package model

import (
	"testing"

	"github.com/xutils/lib-common/xlog"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func TestNewGorm(t *testing.T) {
	xlog.SetupLogDefault()
	defer xlog.Close()
	conf := DbConfig{
		Host:     "127.0.0.1",
		Port:     3306,
		Password: "123456",
		User:     "root",
		Database: "oexpress_data",
		Debug:    true,
		Charset:  "utf8",
	}
	db, err := NewGorm(conf)
	if err != nil {
		xlog.Error("err=%v", err)
		return
	}
	xlog.Info("db=%+v||err=%+v", db, err)
	db.Exec("set time_zone = '+00:00'")
	rows, err := db.Raw("show variables like '%time_zone%'").Rows()
	if err != nil {
		xlog.Error("err=%v", err)
		return
	}
	var name, val string
	for rows.Next() {
		if err = rows.Scan(&name, &val); err != nil {
			panic(err)
		}
		xlog.Info("name=%v||val=%v", name, val)
	}
}
