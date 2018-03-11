package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// SpreadSheetsConfig :
type SpreadSheetsConfig struct {
	Type           string `json:"type"`
	SpreadSheetsID string `json:"spreadsheets_id"`
	SheetName      string `json:"sheet_name"`
	Range          string `json:"range"`
}

// NewJSONSpreadSheetsConfig :
func NewJSONSpreadSheetsConfig(jsonPath string) (*SpreadSheetsConfig, error) {
	var config *SpreadSheetsConfig
	r, err := os.Open(jsonPath)
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(r).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

// Output :
func (c *SpreadSheetsConfig) Output(path string) (string, error) {
	if path == "" {
		path = "resources/config/spreadsheets/" + c.SpreadSheetsID + "_" + c.SheetName + ".json"
	}
	jsonBytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return "", ioutil.WriteFile(path, jsonBytes, 0644)
}
