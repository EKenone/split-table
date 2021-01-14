package split

import (
	"errors"
	"gorm.io/gorm"
	"math"
	"split-table/conf"
	"strings"
)

type SpInterface interface {
	CreateSplitTable() error
	SyncData() error
}

type Split struct {
	DbName string
	Table  string
}

const PartTotal = 1000

// 创建分表
func (s *Split) Create(sp SpInterface) error {
	db := s.db()
	if !db.Migrator().HasTable(s.Table) {
		return errors.New(s.Table + "表不存在")
	}

	return sp.CreateSplitTable()
}

// 同步数据
func (s *Split) SyncData(sp SpInterface) error {
	return sp.SyncData()
}

// 数据库连接
func (s *Split) db() *gorm.DB {
	var db *gorm.DB
	switch s.DbName {
	case conf.CommonDbName:
		db = conf.CommonLink()
	case conf.SaleDbName:
		db = conf.SaleLink()
	}
	return db
}

// 主表的信息
func (s *Split) tableDesc() (map[string]string, int, error) {
	var (
		dataCount float64
		st        []struct {
			Field   string
			Type    string
			Null    string
			Key     string
			Default string
			Extra   string
		}
		fieldType = make(map[string]string)
	)

	db, table := s.db(), s.Table
	db.Raw("SELECT COUNT(*) FROM " + table).Scan(&dataCount)

	if dataCount == 0 {
		return fieldType, 0, errors.New("不需要同步数据")
	}

	db.Debug().Raw("DESC " + table).Scan(&st)

	for _, v := range st {
		ty := "string"
		if strings.ContainsAny(v.Type, "int") {
			ty = "int"
		}
		fieldType[v.Field] = ty
	}

	part := int(math.Ceil(dataCount / PartTotal))

	return fieldType, part, nil
}
