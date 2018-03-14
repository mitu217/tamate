package datasource

import (
	"encoding/csv"
	"errors"
	"os"

	"github.com/Mitu217/tamate/schema"
)

// CSVDatasourceConfig :
type CSVDatasourceConfig struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

// CSVDataSource :
type CSVDataSource struct {
	Config *CSVDatasourceConfig
	Schema *schema.Schema
}

// NewCSVDataSource :
func NewCSVDataSource(config *CSVDatasourceConfig) (*CSVDataSource, error) {
	ds := &CSVDataSource{
		Config: config,
	}
	return ds, nil
}

// GetSchema :
func (ds *CSVDataSource) GetSchema() (*schema.Schema, error) {
	return ds.Schema, nil
}

// SetSchema :
func (ds *CSVDataSource) SetSchema(sc *schema.Schema) error {
	ds.Schema = sc
	return nil
}

// GetRows :
func (ds *CSVDataSource) GetRows() (*Rows, error) {
	r, err := os.Open(ds.Config.Path)
	if err != nil {
		return nil, err
	}

	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, err
	}

	// Get Columns
	columns := make([]string, 0)
	for _, record := range records {
		tagField := record[0]
		if tagField == "COLUMN" {
			sheetColumns := record[1:]
			for _, sheetColumn := range sheetColumns {
				columns = append(columns, sheetColumn)
			}
		}
	}
	if len(columns) == 0 {
		return nil, errors.New("No columns in CSV. Path: " + ds.Config.Path)
	}

	values := append(records[:0], records[1:]...)
	rows := &Rows{
		Columns: columns,
		Values:  values,
	}
	return rows, nil
}

// SetRows :
func (ds *CSVDataSource) SetRows(rows *Rows) error {
	file, err := os.Create(ds.Config.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range append([][]string{rows.Columns}, rows.Values...) {
		err := writer.Write(value)
		if err != nil {
			return err
		}
	}
	return nil
}
