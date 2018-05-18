package handler

import (
	"os"
	"testing"
)

func TestSpannerHandler_GetRows(t *testing.T) {
	dsn := os.Getenv("TAMATE_SPANNER_DSN")
	if dsn == "" {
		t.Skip()
	}

	h, err := NewSpannerHandler(dsn)
	if err != nil {
		t.Fatal(err)
	}

	scs, err := h.GetSchemas()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Schema: %+v", scs[0])
	t.Logf("PK: %+v", scs[0].PrimaryKey)

	rows, err := h.GetRows(scs[0])
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rows.Values)
}