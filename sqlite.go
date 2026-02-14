package sqlite

import (
	tool "github.com/Xiaxiaobaii/autotool"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Sql struct {
	Db      *sql.DB
	SqlChan chan struct {
		f func() error
		c chan error
	}
}

type DataType string
type CountType string
type SqliteType string
type DbTType string

const (
	INT             DataType   = "INT"
	INTEGER         DataType   = "INTEGER"
	INTEGER_PRIMARY DataType   = "INTEGER PRIMARY KEY"
	TINYINT         DataType   = "TINYINT"
	SMALLINT        DataType   = "SMALLINT"
	MEDIUMINT       DataType   = "MEDIUMINT"
	BIGINT          DataType   = "BIGINT"
	INT2            DataType   = "INT2"
	INT8            DataType   = "INT8"
	CHARACTER20     DataType   = "CHARACTER(20)"
	VARCHAR255      DataType   = "VARCHAR(255)"
	NCHAR55         DataType   = "NCHAR(55)"
	NVARCHAR100     DataType   = "NVARCHAR(100)"
	TEXT            DataType   = "TEXT"
	CLOB            DataType   = "CLOB"
	BLOB            DataType   = "BLOB"
	REAL            DataType   = "REAL"
	DOUBLE          DataType   = "DOUBLE"
	FLOAT           DataType   = "FLOAT"
	NUMERIC         DataType   = "NUMERIC"
	DECIMAL105      DataType   = "DECIMAL(10,5)"
	BOOLEAN         DataType   = "BOOLEAN"
	DATE            DataType   = "DATE"
	DATETIME        DataType   = "DATETIME"
	Equal           CountType  = "="
	Lthan           CountType  = "<"
	Gthan           CountType  = ">"
	LEthan          CountType  = "<="
	GEthan          CountType  = ">="
	AND             CountType  = "AND"
	OR              CountType  = "OR"
	NOWHERE         SqliteType = "NoWhere"
	ORREPLACE       DbTType    = "OR REPLACE"
	ORIGNORE        DbTType    = "OR IGNORE"
)

func New(SqliteName string) *Sql {
	db, err := sql.Open("sqlite3", "file:"+tool.Findfile()+SqliteName+"?mode=rwc")
	if err != nil {
		tool.LogPrint("OpenSqliteDataBaseError: "+err.Error(), "Error", 1)
	}
	sql := &Sql{
		Db: db,
		SqlChan: make(chan struct {
			f func() error
			c chan error
		}),
	}

	return sql
}

func New_FormatData(str interface{}) string {
	_, ok := str.(string)
	if ok {
		str = "'" + str.(string) + "'"
	}
	return tool.IntoS(str)
}

func FormatList(args []interface{}) string {
	var ran string
	for _, f := range args {
		typ := dataTypeSwich(f)
		sf := tool.IntoS(f)
		if typ == TEXT || typ == CLOB || typ == NVARCHAR100 {
			sf = fmt.Sprintf("'%s'", sf)
		}
		ran += fmt.Sprintf("%s,", sf)
	}
	ran, _ = strings.CutSuffix(ran, ",")
	return ran
}

func TablesBuild(k []string, v []map[string]DataType) (map[string]map[string]DataType, error) {
	if len(k) != len(v) {
		return nil, tool.Error("The size of k v is not equal.")
	}
	var Re map[string]map[string]DataType = make(map[string]map[string]DataType)
	for i := range k {
		Re[k[i]] = v[i]
	}
	return Re, nil
}

func dataTypeSwich(in interface{}) DataType {
	switch in.(type) {
	case int:
		return INT
	case string:
		return TEXT
	case float64:
		return FLOAT
	case float32:
		return FLOAT
	default:
		return INT
	}
}

var SqlChan chan struct {
	f func() error
	c chan error
} = make(chan struct {
	f func() error
	c chan error
})

func SqlWait(f func() error) error {
	var c chan error = make(chan error)
	SqlChan <- struct {
		f func() error
		c chan error
	}{
		f: f,
		c: c,
	}
	e := <-c
	return e
}

func (sql *Sql) SqlWait(f func() error) error {
	var c chan error = make(chan error)
	sql.SqlChan <- struct {
		f func() error
		c chan error
	}{
		f: f,
		c: c,
	}
	e := <-c
	return e
}

func (sql *Sql) SqlWaiting() {
	for {
		var f = <-sql.SqlChan
		f.c <- f.f()
	}
}

func SqlWaiting() {
	for {
		var f = <-SqlChan
		f.c <- f.f()
	}
}

func init() {
	go SqlWaiting()
}
