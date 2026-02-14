package sqlite

import (
	tool "autotool"
	"encoding/json"
	"fmt"
	"strings"
)

func SplitD(list []string) string {
	result := ""
	for _, v := range list {
		result += fmt.Sprintf("%s,", v)
	}
	result, _ = strings.CutSuffix(result, ",")
	return result
}

func SplitM(hashmap map[string]DataType) string {
	result := ""
	for k, v := range hashmap {
		result += fmt.Sprintf("%s %s,", k, string(v))
	}
	result, _ = strings.CutSuffix(result, ",")
	return result
}

// 创建表
func (sql *Sql) CreateTable(TableName string, kv map[string]DataType) error {
	query := fmt.Sprintf("CREATE TABLE %s (%s);", TableName, SplitM(kv))
	_, err := sql.Db.Exec(query)
	return err
}

// 补丁，CREATE TABLE IF NOT EXISTS
func (sql *Sql) CreateIfNotTable(TableName string, kv map[string]DataType) error {
	return sql.CreateTable("IF NOT EXISTS "+TableName, kv)
}

// 检查表是否存在
func (sql *Sql) CheckTable(TableName string) bool {
	query := fmt.Sprintf("SELECT * FROM %s;", TableName)
	_, err := sql.Db.Exec(query)
	return err == nil
}

// 删除表
func (sql *Sql) DeleteTable(TableName string) error {
	query := fmt.Sprintf("DROP TABLE %s;", TableName)
	_, err := sql.Db.Exec(query)
	return err
}

// 插入数据
func (sql *Sql) Insert(TableName string, kv map[string]any, args ...DbTType) error {
	query := "INSERT INTO " + TableName + "("
	if len(args) > 0 {
		switch args[0] {
		case ORIGNORE:
			query = "INSERT OR IGNORE INTO " + TableName + "("
		case ORREPLACE:
			query = "INSERT OR REPLACE INTO " + TableName + "("
		}
	}
	var Datas []any = make([]any, 0)
	for k, v := range kv {
		query = query + k + ","
		Datas = append(Datas, v)
	}
	query, _ = strings.CutSuffix(query, ",")
	query = query + ") VALUES ("
	query = query + FormatList(Datas) + ");"
	_, e := sql.Db.Exec(query)
	return e
}

func (sql *Sql) AlterTable(TableName string, DataName string, Typ DataType) error {
	//ALTER TABLE 表名 ADD COLUMN 列名 数据类型;
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", TableName, DataName, string(Typ))
	_, e := sql.Db.Exec(query)
	return e
}

// 批量表创建
func (sql *Sql) CreateTables(Tables map[string]map[string]DataType) error {
	var err error
	for k, v := range Tables {
		if e := sql.CreateIfNotTable(k, v); e != nil {
			err = tool.Error("CreateTable " + k + " Error: " + e.Error())
		}
	}
	return err
}

func (sql *Sql) ForceCreateTableData(Tables map[string]map[string]DataType) error {
	var err error
	for k, v := range Tables {
		if e := sql.CreateTable(k, v); e != nil {
			if strings.Contains(e.Error(), "already exists") {
				for ik, iv := range v {
					sql.AlterTable(k, ik, iv)
				}
			} else {
				err = tool.Error("CreateTable " + k + " Error: " + e.Error())
			}
		}
	}
	return err
}

// @description 搜索数据

// @param TableName 搜索表名, Selectis 搜索字段(禁止*)，Whereis Where语句, args按搜索字段循序填写回传值
func (sql *Sql) SelectOne(TableName, Selectis, Whereis string, args ...any) error {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ;", Selectis, TableName, Whereis)
	if Whereis == "NoWhere" {
		query = fmt.Sprintf("SELECT %s FROM %s ;", Selectis, TableName)
	}

	// 弃用, 仅兼容
	if strings.Contains(Whereis, " PASSW()") {
		tool.Dsend(&Whereis, 8)
		query = "SELECT " + Selectis + " FROM " + TableName + " " + Whereis + " ;"
	}

	q, e := sql.Db.Query(query)
	if e != nil {
		return e
	}
	defer q.Close()
	Success := false
	for q.Next() {
		Success = true
		q.Scan(args...)
	}
	if Success {
		return nil
	}
	return tool.Error("No Data")
}

// TableName: 表名
// Selectis: 搜索字段列表
// Whereis: WHERE后跟语句,""则为无Where语句
// ret: 传入一个*[]any参数, 将返回结果填充进数组中, 数组内容为map[字段名]字段内容

func (sql *Sql) Selects(TableName, Selectis, Whereis string, ret *[]any) error {
	query := "SELECT " + Selectis + " FROM " + TableName + " WHERE " + Whereis + " ;"
	if Whereis == "" {
		query = "SELECT " + Selectis + " FROM " + TableName + ";"
	}
	arraystring := []string{}
	for v := range strings.SplitSeq(Selectis, ",") {
		arraystring = append(arraystring, strings.Trim(v, " "))
	}
	arraylen := len(arraystring)
	q, e := sql.Db.Query(query)
	if e != nil {
		return e
	}
	defer q.Close()
	for q.Next() {
		arg := make([]any, 0)
		for i := 0; arraylen > i; i++ {
			arg = append(arg, new(any))
		}
		q.Scan(arg...)
		var temp_array = map[string]any{}
		for i := 0; arraylen > i; i++ {
			temp_array[arraystring[i]] = *arg[i].(*any)
		}
		*ret = append(*ret, temp_array)
	}
	return nil
}

func (sql *Sql) SelectArg(TableName, Selectis, Whereis string, args []any) error {
	return sql.SelectOne(TableName, Selectis, Whereis, args...)
}

func (sql *Sql) RandomData(TableName, Selectis, Whereis string, args []any) error {
	if Whereis == "" || Whereis == " " {
		return sql.SelectOne(TableName, Selectis, "ORDER BY RANDOM() limit 1 PASSW()", args...)
	} else {
		return sql.SelectOne(TableName, Selectis, "WHERE "+Whereis+" ORDER BY RANDOM() limit 1 PASSW()", args...)
	}

}

func (sql *Sql) Delete(TableName, Whereis string) error {
	// DELETE FROM COMPANY WHERE ID = 7;
	query := fmt.Sprintf("DELETE FROM %s WHERE %s;", TableName, Whereis)
	if Whereis == "" {
		query = fmt.Sprintf("DELETE FROM %s;", TableName)
	}
	_, e := sql.Db.Exec(query)
	return e
}

func (sql *Sql) CheckData(TableName string, Whereis string) bool {
	query := "Select * FROM " + TableName + " WHERE " + Whereis + ";"
	if Whereis == "" {
		query = fmt.Sprintf("Select * FROM %s;", TableName)
	}
	row, err := sql.Db.Query(query)
	if err != nil {
		return false
	}
	defer row.Close()
	if row.Next() {
		return true
	}
	return false
}

func (sql *Sql) UpdateData(TableName string, Updateis map[string]any, Whereis string) error {
	UpdateQuery := ""
	for k, v := range Updateis {
		UpdateQuery += fmt.Sprintf("%s = %s,", k, New_FormatData(v))
	}
	UpdateQuery = strings.Trim(UpdateQuery, ",")
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s ;", TableName, UpdateQuery, Whereis)
	_, e := sql.Db.Exec(query)
	return e
}

// TableName:表名，Id:ID键名，Ids:ID键值，Plus:Plus键名, p:加数
func (sql *Sql) Updateplus(TableName string, Id string, Ids string, Plus string, p int) error {
	// INSERT INTO vocabulary(word) VALUES('jovial')
	// ON CONFLICT(word) DO UPDATE SET count=count+1;
	Query := fmt.Sprintf("INSERT INTO %s(%s,%s) VALUES(%s, 1) ON CONFLICT(%s) DO UPDATE SET %s=%s+%d;", TableName, Id, Plus, Ids, Id, Plus, Plus, p)
	_, e := sql.Db.Exec(Query)
	return e
}

func (sql *Sql) InsertData(TableName string, Insertis map[string]any) error {
	InsertKey := ""
	var Datas []any = make([]any, 0)
	for k, v := range Insertis {
		InsertKey += k + ","
		Datas = append(Datas, v)
	}
	InsertKey = strings.Trim(InsertKey, ",")
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s );", TableName, InsertKey, FormatList(Datas))
	_, e := sql.Db.Exec(query)
	return e
}

func (sql *Sql) InsertData_Struct(TableName string, Updateis any) error {
	rawData, e := json.Marshal(Updateis)
	if e != nil {
		return e
	}
	var Data map[string]any
	e = json.Unmarshal(rawData, &Data)
	if e != nil {
		return e
	}
	return sql.InsertData(TableName, Data)
}

func (sql *Sql) UpdateData_Struct(TableName string, Updateis any, Whereis string) error {
	rawData, e := json.Marshal(Updateis)
	if e != nil {
		return e
	}
	var Data map[string]any
	e = json.Unmarshal(rawData, &Data)
	if e != nil {
		return e
	}
	return sql.UpdateData(TableName, Data, Whereis)
}
