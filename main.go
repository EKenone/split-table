package main

import (
	"flag"
	"log"
	"split-table/conf"
	"split-table/split"
)

var (
	dbName   string
	table    string
	modField string
	million  int
)

func init() {
	flag.StringVar(&dbName, "db", "common", "操作的数据库")
	flag.StringVar(&table, "t", "", "操作的表名")
	flag.StringVar(&modField, "mf", "", "分表的取模字段，多字段的时候逗号分隔")
	flag.IntVar(&million, "ml", 0, "分表数据的数据量的数据百万条数")
}

func main() {
	log.Println("开始")
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Panic(err)
	}

	s := split.Split{
		DbName: dbName,
		Table:  table,
	}

	sp := split.FieldMod{
		Split:    s,
		ModField: modField,
	}

	err := s.Create(&sp)
	if err != nil {
		log.Panicln(err)
	}

	err = s.SyncData(&sp)
	if err != nil {
		log.Panicln(err)
	}
	log.Println("结束")
}
