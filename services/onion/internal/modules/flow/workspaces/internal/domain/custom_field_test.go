package domain

import (
	"errors"
	"strings"
	"testing"
)

func selectField() CustomField {
	return CustomField{ID: 7, Name: "Severity", Type: FieldTypeSelect, Options: []FieldOption{
		{ID: "a1", Label: "Low"}, {ID: "b2", Label: "High"},
	}}
}

func TestNormalizeFieldInput_AssignsOptionIDsAndResolvesDefaultLabel(t *testing.T) {
	in, removed, err := NormalizeFieldInput(CustomFieldInput{
		Name: "  Severity ", Type: FieldTypeSelect,
		Options: []FieldOption{{Label: "Low"}, {Label: "High", Color: "#ff0000"}},
		Default: FieldValue{Value: "high"},
	}, nil)
	if err != nil {
		t.Fatalf("NormalizeFieldInput: %v", err)
	}
	if in.Name != "Severity" || len(removed) != 0 {
		t.Fatalf("name %q removed %v", in.Name, removed)
	}
	if in.Options[0].ID == "" || in.Options[0].Color != DefaultFieldOptionColor || in.Options[1].Color != "#ff0000" {
		t.Fatalf("options = %+v", in.Options)
	}
	if in.Default.Value != in.Options[1].ID {
		t.Fatalf("default %q, want the id of High %q", in.Default.Value, in.Options[1].ID)
	}
}

func TestNormalizeFieldInput_Rejects(t *testing.T) {
	cases := map[string]struct {
		in   CustomFieldInput
		want error
	}{
		"no name":            {CustomFieldInput{Name: " ", Type: FieldTypeText}, ErrFieldNameRequired},
		"bad type":           {CustomFieldInput{Name: "x", Type: "rating"}, ErrInvalidFieldType},
		"select no options":  {CustomFieldInput{Name: "x", Type: FieldTypeSelect}, ErrFieldOptionsRequired},
		"text with options":  {CustomFieldInput{Name: "x", Type: FieldTypeText, Options: []FieldOption{{Label: "a"}}}, ErrFieldOptionsNotAllowed},
		"duplicate label":    {CustomFieldInput{Name: "x", Type: FieldTypeSelect, Options: []FieldOption{{Label: "A"}, {Label: "a"}}}, ErrFieldOptionDuplicate},
		"foreign option id":  {CustomFieldInput{Name: "x", Type: FieldTypeSelect, Options: []FieldOption{{ID: "zz", Label: "A"}}}, ErrFieldOptionUnknown},
		"name over 64 chars": {CustomFieldInput{Name: strings.Repeat("x", MaxFieldNameLength+1), Type: FieldTypeText}, ErrFieldNameTooLong},
	}
	for name, c := range cases {
		if _, _, err := NormalizeFieldInput(c.in, nil); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", name, err, c.want)
		}
	}
}

func TestNormalizeFieldInput_ReportsRemovedOptions(t *testing.T) {
	existing := selectField().Options
	in, removed, err := NormalizeFieldInput(CustomFieldInput{
		Name: "Severity", Type: FieldTypeSelect,
		Options: []FieldOption{{ID: "b2", Label: "Critical"}, {Label: "Medium"}},
	}, existing)
	if err != nil {
		t.Fatalf("NormalizeFieldInput: %v", err)
	}
	if len(removed) != 1 || removed[0] != "a1" {
		t.Fatalf("removed = %v, want [a1]", removed)
	}
	if in.Options[0].ID != "b2" || in.Options[0].Label != "Critical" {
		t.Fatalf("renamed option lost its id: %+v", in.Options[0])
	}
}

func TestNormalizeFieldValue(t *testing.T) {
	cases := []struct {
		name  string
		field CustomField
		in    FieldValue
		want  FieldValue
		code  string
	}{
		{"number canonical", CustomField{Type: FieldTypeNumber}, FieldValue{Value: "12.50"}, FieldValue{Value: "12.5"}, ""},
		{"number bad", CustomField{Type: FieldTypeNumber}, FieldValue{Value: "1e3"}, FieldValue{}, FieldErrNumber},
		{"date", CustomField{Type: FieldTypeDate}, FieldValue{Value: "2026-11-19"}, FieldValue{Value: "2026-11-19"}, ""},
		{"date bad", CustomField{Type: FieldTypeDate}, FieldValue{Value: "19/11/2026"}, FieldValue{}, FieldErrDate},
		{"url", CustomField{Type: FieldTypeURL}, FieldValue{Value: "https://rtylr.com"}, FieldValue{Value: "https://rtylr.com"}, ""},
		{"url scheme", CustomField{Type: FieldTypeURL}, FieldValue{Value: "javascript:alert(1)"}, FieldValue{}, FieldErrURL},
		{"checkbox", CustomField{Type: FieldTypeCheckbox}, FieldValue{Value: "true"}, FieldValue{Value: "true"}, ""},
		{"checkbox bad", CustomField{Type: FieldTypeCheckbox}, FieldValue{Value: "yes"}, FieldValue{}, FieldErrCheckbox},
		{"select", selectField(), FieldValue{Value: "a1"}, FieldValue{Value: "a1"}, ""},
		{"select unknown", selectField(), FieldValue{Value: "Low"}, FieldValue{}, FieldErrOption},
		{"multi dedupes", CustomField{Type: FieldTypeMultiSelect, Options: selectField().Options}, FieldValue{Values: []string{"a1", "a1", "b2"}}, FieldValue{Values: []string{"a1", "b2"}}, ""},
		{"multi given value", CustomField{Type: FieldTypeMultiSelect, Options: selectField().Options}, FieldValue{Value: "a1"}, FieldValue{}, FieldErrShape},
		{"person ids", CustomField{Type: FieldTypePerson}, FieldValue{Values: []string{"007"}}, FieldValue{Values: []string{"7"}}, ""},
		{"person bad", CustomField{Type: FieldTypePerson}, FieldValue{Values: []string{"bob"}}, FieldValue{}, FieldErrPerson},
		{"empty clears", CustomField{Type: FieldTypeText}, FieldValue{Value: "  "}, FieldValue{}, ""},
	}
	for _, c := range cases {
		got, err := NormalizeFieldValue(c.field, c.in)
		if c.code != "" {
			var fe *FieldValueError
			if !errors.As(err, &fe) || fe.Code != c.code {
				t.Errorf("%s: err = %v, want code %s", c.name, err, c.code)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if got.Value != c.want.Value || len(got.Values) != len(c.want.Values) {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
			continue
		}
		for i := range got.Values {
			if got.Values[i] != c.want.Values[i] {
				t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
			}
		}
	}
}

func TestBuildTaskFieldFilter(t *testing.T) {
	if _, err := BuildTaskFieldFilter(selectField(), FieldOpAny, []string{"a1", "b2"}); err != nil {
		t.Errorf("select any: %v", err)
	}
	if _, err := BuildTaskFieldFilter(selectField(), FieldOpAny, []string{"nope"}); !errors.Is(err, ErrFieldFilterInvalid) {
		t.Errorf("select any unknown option err = %v", err)
	}
	if _, err := BuildTaskFieldFilter(CustomField{Type: FieldTypeText}, FieldOpRange, []string{"a", "b"}); !errors.Is(err, ErrFieldFilterInvalid) {
		t.Errorf("text range err = %v", err)
	}
	f, err := BuildTaskFieldFilter(CustomField{Type: FieldTypeNumber}, FieldOpRange, []string{"", "10.50"})
	if err != nil || f.Values[0] != "" || f.Values[1] != "10.5" {
		t.Errorf("open range = %+v, %v", f, err)
	}
	if _, err := BuildTaskFieldFilter(CustomField{Type: FieldTypeMultiSelect, Options: selectField().Options}, FieldOpEq, []string{"a1"}); !errors.Is(err, ErrFieldFilterInvalid) {
		t.Errorf("multi eq err = %v", err)
	}
}

func TestParseFieldSort(t *testing.T) {
	if id, ok, err := ParseFieldSort("field:9"); id != 9 || !ok || err != nil {
		t.Errorf("field sort = %d %v %v", id, ok, err)
	}
	if _, ok, err := ParseFieldSort("due_date"); ok || err != nil {
		t.Errorf("plain sort ok=%v err=%v", ok, err)
	}
	if _, _, err := ParseFieldSort("field:x"); !errors.Is(err, ErrFieldSortInvalid) {
		t.Errorf("bad field sort err = %v", err)
	}
	if FieldSortable(FieldTypePerson) || !FieldSortable(FieldTypeSelect) {
		t.Errorf("FieldSortable wrong")
	}
}

func TestValidateFieldOrder(t *testing.T) {
	if err := ValidateFieldOrder([]int64{1, 2, 3}, []int64{3, 1, 2}); err != nil {
		t.Errorf("valid order: %v", err)
	}
	for name, ids := range map[string][]int64{"missing": {1, 2}, "dup": {1, 1, 2}, "foreign": {1, 2, 4}} {
		if err := ValidateFieldOrder([]int64{1, 2, 3}, ids); !errors.Is(err, ErrFieldOrderMismatch) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
