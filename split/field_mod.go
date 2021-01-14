package split

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type FieldMod struct {
	Split
	ModField string
}

// 创建分表
func (f *FieldMod) CreateSplitTable() error {
	var (
		cSql string
		t    string
	)
	db, table := f.db(), f.Table
	res := db.Raw("SHOW CREATE TABLE " + table).Row()
	if err := res.Scan(&t, &cSql); err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		newTable := table + "_" + strconv.Itoa(i)
		if db.Migrator().HasTable(newTable) {
			log.Println("已存在表" + newTable)
			continue
		}
		sql := strings.Replace(cSql, table, newTable, 1)
		sql = strings.Replace(sql, "AUTO_INCREMENT", "", 1)
		err := db.Exec(sql).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// 同步数据
func (f *FieldMod) SyncData() error {
	fieldType, part, err := f.tableDesc()

	if err != nil {
		return err
	}

	db, table := f.db(), f.Table

	for i := 0; i < part; i++ {
		var results []map[string]interface{}
		of := i * PartTotal
		db.Table(table).Offset(of).Limit(PartTotal).Order("id").Find(&results)
		f.insertData(results, fieldType)
	}

	return nil
}

// 添加数据到各个分表
func (f *FieldMod) insertData(results []map[string]interface{}, fieldType map[string]string) {
	table, saveData := f.Table, f.modData(results)
	g := sync.WaitGroup{}
	for m, rows := range saveData {
		rows, m := rows, m
		g.Add(1)
		go func() {
			defer g.Done()
			db := f.db()
			rowData := rows[0].(map[string]interface{})
			f, l := insertFields(rowData)
			var v []string
			for _, row := range rows {
				one := row.(map[string]interface{})
				str := "("
				for i := 0; i < l; i++ {
					key := strings.Trim(f[i], "`")
					if fieldType[f[i]] == "int" {
						str += fmt.Sprintf("%v,", one[key])
					} else {
						str += fmt.Sprintf("'%v',", one[key])
					}
				}
				str = strings.TrimRight(str, ",")
				str += ")"
				v = append(v, str)
			}
			fix := m
			it := table + "_" + strconv.Itoa(fix)

			isql := "INSERT INTO `" + it + "` (" + strings.Join(f, ",") + ") VALUES " + strings.Join(v, ",") + ";"
			db.Exec(isql)
		}()
	}
	g.Wait()
}

// 把数据按取模区分好
func (f *FieldMod) modData(results []map[string]interface{}) map[int][]interface{} {
	saveData := make(map[int][]interface{})
	fm := strings.Split(f.ModField, ",")
	for _, result := range results {
		var v int64
		for _, m := range fm {
			key := ""
			switch reflect.TypeOf(result[m]).Kind() {
			case reflect.Uint32:
				key = strconv.Itoa(int(result[m].(uint32)))
			default:
				key = result[m].(string)
			}
			val, _ := strconv.ParseInt(key, 10, 64)
			v = v + val
		}
		m := v % 10
		saveData[int(m)] = append(saveData[int(m)], result)
	}
	return saveData
}

// 添加的字段和字段的长度
func insertFields(row map[string]interface{}) ([]string, int) {
	f := make([]string, 0)
	for k, _ := range row {
		key := "`" + k + "`"
		f = append(f, key)
	}
	l := len(f)

	return f, l
}
