package queries

import (
	"context"
	"testing"
)

func TestQueryBuilderIfNotZero(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	builder := NewBuilder("TestQueryBuilderIfNotZero")
	builder.SetIfNotZero(ctx, "a", 1)
	builder.SetIfNotZero(ctx, "b", 0)
	builder.SetIfNotZero(ctx, "c", "hello")
	builder.SetIfNotZero(ctx, "d", "")
	builder.SetIfNotZero(ctx, "e", true)
	builder.SetIfNotZero(ctx, "f", false)
	builder.SetIfNotZero(ctx, "g", 1.0)
	builder.SetIfNotZero(ctx, "h", 0.0)

	if len(builder.columns) != 4 {
		t.Errorf("len(builder.columns) = %v, want 4", len(builder.columns))
	}
	if len(builder.values) != 4 {
		t.Errorf("len(builder.values) = %v, want 4", len(builder.values))
	}
	if builder.columns[0] != "a" {
		t.Errorf("builder.columns[0] = %v, want %v", builder.columns[0], "a")
	}
	if builder.columns[1] != "c" {
		t.Errorf("builder.columns[1] = %v, want %v", builder.columns[1], "c")
	}
	if builder.columns[2] != "e" {
		t.Errorf("builder.columns[2] = %v, want %v", builder.columns[2], "e")
	}
	if builder.columns[3] != "g" {
		t.Errorf("builder.columns[3] = %v, want %v", builder.columns[3], "g")
	}
	if builder.values[0] != 1 {
		t.Errorf("builder.values[0] = %v, want %v", builder.values[0], 1)
	}
	if builder.values[1] != "hello" {
		t.Errorf("builder.values[1] = %v, want %v", builder.values[1], "hello")
	}
	if builder.values[2] != true {
		t.Errorf("builder.values[2] = %v, want %v", builder.values[2], true)
	}
	if builder.values[3] != 1.0 {
		t.Errorf("builder.values[3] = %v, want %v", builder.values[3], 1.0)
	}
}

func TestQueryBuilderIf(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	builder := NewBuilder("TestQueryBuilderIf")
	builder.SetIf(ctx, "a", 1, true)
	builder.SetIf(ctx, "b", 0, false)
	builder.SetIf(ctx, "c", "hello", true)
	builder.SetIf(ctx, "d", "", false)
	builder.SetIf(ctx, "e", true, true)
	builder.SetIf(ctx, "f", false, false)
	builder.SetIf(ctx, "g", 1.0, true)
	builder.SetIf(ctx, "h", 0.0, false)

	if len(builder.columns) != 4 {
		t.Errorf("len(builder.columns) = %v, want 4", len(builder.columns))
	}
	if len(builder.values) != 4 {
		t.Errorf("len(builder.values) = %v, want 4", len(builder.values))
	}
	if builder.columns[0] != "a" {
		t.Errorf("builder.columns[0] = %v, want %v", builder.columns[0], "a")
	}
	if builder.columns[1] != "c" {
		t.Errorf("builder.columns[1] = %v, want %v", builder.columns[1], "c")
	}
	if builder.columns[2] != "e" {
		t.Errorf("builder.columns[2] = %v, want %v", builder.columns[2], "e")
	}
	if builder.columns[3] != "g" {
		t.Errorf("builder.columns[3] = %v, want %v", builder.columns[3], "g")
	}
	if builder.values[0] != 1 {
		t.Errorf("builder.values[0] = %v, want %v", builder.values[0], 1)
	}
	if builder.values[1] != "hello" {
		t.Errorf("builder.values[1] = %v, want %v", builder.values[1], "hello")
	}
	if builder.values[2] != true {
		t.Errorf("builder.values[2] = %v, want %v", builder.values[2], true)
	}
	if builder.values[3] != 1.0 {
		t.Errorf("builder.values[3] = %v, want %v", builder.values[3], 1.0)
	}
}

func TestBuilder_Select(t *testing.T) {
	t.Parallel()
	builder := NewBuilder("TestBuilder_Select")
	builder.OrderBy = "id DESC"
	builder.Limit = 50
	builder.Offset = 100

	query, _ := builder.Select("some_table", "col1,col2,col3")

	expectedQuery := "-- :TestBuilder_Select:\nSELECT col1,col2,col3 FROM some_table ORDER BY id DESC LIMIT 50 OFFSET 100"

	if query != expectedQuery {
		t.Errorf("query = %v, want %v", query, expectedQuery)
	}
}

func TestBuilder_Insert(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	builder := NewBuilder("TestBuilder_Insert")
	builder.SetIfNotZero(ctx, "some_column", "some_value")
	query, _ := builder.Insert("some_table")

	expectedQuery := "-- :TestBuilder_Insert:\nINSERT INTO some_table (some_column) VALUES ($1)"

	if query != expectedQuery {
		t.Errorf("query = %v, want %v", query, expectedQuery)
	}
}
