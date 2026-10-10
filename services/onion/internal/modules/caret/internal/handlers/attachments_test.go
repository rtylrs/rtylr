package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestStreamRequestCarriesTextFiles(t *testing.T) {
	var req GenerateTextStreamRequest
	if err := json.Unmarshal([]byte(`{"prompt":"sum it","user_id":1,"files":[{"name":"sales.csv","text":"a,b\n1,2"}]}`), &req); err != nil {
		t.Fatal(err)
	}
	if len(req.Files) != 1 || req.Files[0].Name != "sales.csv" {
		t.Fatalf("files = %#v", req.Files)
	}
	handler := &GenerateTextStreamHandler{}
	if err := handler.Validate(context.Background(), &req); err != nil {
		t.Fatal(err)
	}
	req.Files = append(req.Files, req.Files[0], req.Files[0], req.Files[0])
	if err := handler.Validate(context.Background(), &req); err == nil || !strings.Contains(err.Error(), "up to 3 files") {
		t.Fatalf("four files: %v", err)
	}
}
