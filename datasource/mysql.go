package datasource

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"context"
	"github.com/go-sql-driver/mysql"
	"reflect"
	"time"
)

type MySQLDatasource struct {
	DSN string `json:"dsn"`
	db  *sql.DB
}

// NewMySQLDatasource is create MySQLDatasource instance method
func NewMySQLDatasource(dsn string) (*MySQLDatasource, error) {
	return &MySQLDatasource{
		DSN: dsn,
	}, nil
}

// Open is call by datasource when create instance
func (h *MySQLDatasource) Open() error {
	if h.db == nil {
		db, err := sql.Open("mysql", h.DSN)
		if err != nil {
			return err
		}
		if err := db.Ping(); err != nil {
			return err
		}
		h.db = db
	}
	return nil
}

// Close is call by datasource when free instance
func (h *MySQLDatasource) Close() error {
	if h.db != nil {
		err := h.db.Close()
		h.db = nil
		if err != nil {
			return err
		}
	}
	return nil
}

// GetSchemas is get all schemas method
func (h *MySQLDatasource) createAllSchemaMap() (map[string]*Schema, error) {
	// get schemas
	sqlRows, err := h.db.Query("SELECT TABLE_NAME, COLUMN_NAME, ORDINAL_POSITION, COLUMN_TYPE, COLUMN_KEY, IS_NULLABLE, EXTRA FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE()")
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	// scan results
	schemaMap := make(map[string]*Schema)
	for sqlRows.Next() {
		var tableName string
		var columnName string
		var ordinalPosition int
		var columnType string
		var columnKey string
		var isNullable string
		var extra string
		if err := sqlRows.Scan(&tableName, &columnName, &ordinalPosition, &columnType, &columnKey, &isNullable, &extra); err != nil {
			return nil, err
		}
		// prepare schema
		if _, ok := schemaMap[tableName]; !ok {
			schemaMap[tableName] = &Schema{Name: tableName}
		}
		schema := schemaMap[tableName]
		// set column in schema
		if strings.Contains(columnKey, "PRI") {
			if schema.PrimaryKey == nil {
				schema.PrimaryKey = &PrimaryKey{}
			}
			schema.PrimaryKey.ColumnNames = append(schema.PrimaryKey.ColumnNames, columnName)
		}
		valueType, err := mysqlColumnTypeToValueType(columnType)
		if err != nil {
			return nil, err
		}
		column := &Column{
			Name:            columnName,
			OrdinalPosition: ordinalPosition - 1,
			Type:            valueType,
			NotNull:         isNullable != "YES",
			AutoIncrement:   strings.Contains(extra, "auto_increment"),
		}
		schema.Columns = append(schema.Columns, column)
		schemaMap[tableName] = schema
	}

	return schemaMap, nil
}

// TODO: various MySQL types support
func mysqlColumnTypeToValueType(ct string) (ColumnType, error) {
	ct = strings.ToLower(ct)
	if strings.HasPrefix(ct, "int") ||
		strings.HasPrefix(ct, "smallint") ||
		strings.HasPrefix(ct, "mediumint") ||
		strings.HasPrefix(ct, "bigint") {
		return ColumnTypeInt, nil
	}
	if strings.HasPrefix(ct, "float") ||
		strings.HasPrefix(ct, "double") ||
		strings.HasPrefix(ct, "decimal") {
		return ColumnTypeFloat, nil
	}
	if strings.HasPrefix(ct, "char") ||
		strings.HasPrefix(ct, "varchar") ||
		strings.HasPrefix(ct, "text") ||
		strings.HasPrefix(ct, "mediumtext") ||
		strings.HasPrefix(ct, "longtext") ||
		strings.HasPrefix(ct, "json") {
		return ColumnTypeString, nil
	}
	if strings.HasPrefix(ct, "datetime") ||
		strings.HasPrefix(ct, "timestamp") {
		return ColumnTypeDatetime, nil
	}
	if strings.HasPrefix(ct, "date") {
		return ColumnTypeDate, nil
	}
	if strings.HasPrefix(ct, "blob") {
		return ColumnTypeBytes, nil
	}
	return ColumnTypeNull, fmt.Errorf("convertion not found for MySQL type: %s", ct)
}

func (h *MySQLDatasource) GetAllSchema(ctx context.Context) ([]*Schema, error) {
	allMap, err := h.createAllSchemaMap()
	if err != nil {
		return nil, err
	}

	var all []*Schema
	for _, sc := range allMap {
		all = append(all, sc)
	}
	return all, nil
}

// GetSchema is get schema method
func (h *MySQLDatasource) GetSchema(ctx context.Context, name string) (*Schema, error) {
	all, err := h.createAllSchemaMap()
	if err != nil {
		return nil, err
	}
	for scName, sc := range all {
		if scName == name {
			return sc, nil
		}
	}
	return nil, errors.New("schema not found: " + name)
}

// SetSchema is set schema method
func (h *MySQLDatasource) SetSchema(ctx context.Context, schema *Schema) error {
	return errors.New("not support SetSchema()")
}

// GetRows is get rows method
func (h *MySQLDatasource) GetRows(ctx context.Context, schema *Schema) ([]*Row, error) {
	// get data
	sqlRows, err := h.db.Query(fmt.Sprintf("SELECT * FROM %s", schema.Name))
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	var rows []*Row
	for sqlRows.Next() {
		row, err := scanRow(sqlRows, schema)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func scanRow(sqlRows *sql.Rows, sc *Schema) (*Row, error) {
	rowValues := make(RowValues)
	ptrs := make([]interface{}, len(sc.Columns))
	i := 0
	for _, col := range sc.Columns {
		dvp := reflect.New(colToMySQLType(col)).Interface()
		ptrs[i] = dvp
		i++
	}
	if err := sqlRows.Scan(ptrs...); err != nil {
		return nil, err
	}
	i = 0
	for _, col := range sc.Columns {
		v := reflect.ValueOf(ptrs[i]).Elem().Interface()
		rowValues[col.Name] = &GenericColumnValue{Column: col, Value: v}
		i++
	}
	return &Row{Values: rowValues}, nil
}

func colToMySQLType(c *Column) reflect.Type {
	switch c.Type {
	case ColumnTypeInt:
		if c.NotNull {
			return reflect.TypeOf(int64(0))
		}
		return reflect.TypeOf(sql.NullInt64{})

	case ColumnTypeFloat:
		if c.NotNull {
			return reflect.TypeOf(float64(0))
		}
		return reflect.TypeOf(sql.NullFloat64{})
	case ColumnTypeBool:
		if c.NotNull {
			return reflect.TypeOf(false)
		}
		return reflect.TypeOf(sql.NullBool{})
	case ColumnTypeDatetime, ColumnTypeDate:
		if c.NotNull {
			return reflect.TypeOf(time.Time{})
		}
		return reflect.TypeOf(mysql.NullTime{})
	case ColumnTypeString:
		if c.NotNull {
			return reflect.TypeOf("")
		}
		return reflect.TypeOf(sql.NullString{})
	case ColumnTypeBytes:
		return reflect.TypeOf([]byte{})
	}
	return reflect.TypeOf(nil)
}

// SetRows is set rows method
func (h *MySQLDatasource) SetRows(ctx context.Context, schema *Schema, rows []*Row) error {
	return errors.New("MySQLDatasource does not support SetRows()")
}
