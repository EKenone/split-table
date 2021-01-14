package conf

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

type Link struct {
	Db   *gorm.DB
	Once sync.Once
}

var (
	common = Link{}
	sale   = Link{}
)

const (
	CommonDbName = "common"
	SaleDbName   = "sale"
)

// common配置的数据库连接
func CommonLink() *gorm.DB {
	common.Once.Do(func() {
		common.Db = newLink(CommonDbName)
	})
	return common.Db
}

// sale配置的数据库连接
func SaleLink() *gorm.DB {
	sale.Once.Do(func() {
		sale.Db = newLink(SaleDbName)
	})
	return sale.Db
}

// 数据库连接实例化
func newLink(dbName string) *gorm.DB {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second * 3,
			Colorful:      false,
			LogLevel:      logger.Silent,
		},
	)

	db, err := gorm.Open(mysql.Open(getDsn(dbName)), &gorm.Config{
		PrepareStmt: true,
		Logger:      newLogger,
	})

	if err != nil {
		log.Println(err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour * 2)

	return db
}

// 数据库的dsn
func getDsn(dbName string) string {
	dsn := ""
	switch dbName {
	case CommonDbName:
		dsn = Conf.Mysql.Common.User + ":" + Conf.Mysql.Common.Pass + "@tcp(" + Conf.Mysql.Common.Host + ":" + Conf.Mysql.Common.Port + ")/" + Conf.Mysql.Common.Db
	case SaleDbName:
		dsn = Conf.Mysql.Sale.User + ":" + Conf.Mysql.Sale.Pass + "@tcp(" + Conf.Mysql.Sale.Host + ":" + Conf.Mysql.Sale.Port + ")/" + Conf.Mysql.Sale.Db
	}

	if dsn == "" {
		log.Panicln("数据库DSN不存在")
	}

	return dsn + "?charset=utf8&loc=Asia%2FShanghai&parseTime=true"
}
