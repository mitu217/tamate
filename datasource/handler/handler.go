package handler

// Column is table column
type Column struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	NotNull       bool   `json:"not_null"`
	AutoIncrement bool   `json:"auto_increment"`
}

// Rows is table records
type Rows struct {
	Values [][]string `json:"values"`
}

// Schema is column definitions at table
type Schema struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

// ColumnNames is get column names from schema columns
func (sc *Schema) ColumnNames() []string {
	columnNames := make([]string, len(sc.Columns))
	for i := range sc.Columns {
		columnNames[i] = sc.Columns[i].Name
	}
	return columnNames
}

// Handler is read and write datasource interface
type Handler interface {
	Open() error
	Close() error
	GetSchemas() (*[]Schema, error)
	GetSchema(*Schema) error
	SetSchema(*Schema) error
	GetRows(*Schema) (*Rows, error)
	SetRows(*Schema, *Rows) error
}