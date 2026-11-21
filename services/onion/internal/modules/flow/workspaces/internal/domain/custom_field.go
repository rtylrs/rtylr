package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Custom fields (T-580): a project defines fields, a task holds one value per
// field. Everything in this file is pure so the rules can be tested without a
// database.

const (
	FieldTypeText        = "text"
	FieldTypeNumber      = "number"
	FieldTypeDate        = "date"
	FieldTypeSelect      = "select"
	FieldTypeMultiSelect = "multi_select"
	FieldTypePerson      = "person"
	FieldTypeURL         = "url"
	FieldTypeCheckbox    = "checkbox"

	MaxFieldsPerProject   = 50
	MaxFieldOptions       = 100
	MaxFieldNameLength    = 64
	MaxFieldOptionLabel   = 64
	MaxFieldTextLength    = 2000
	MaxFieldPeople        = 20
	MaxFieldValuesPerCall = 50
	MaxFieldFilterClauses = 10

	DefaultFieldOptionColor = "#6366f1"
	// FieldOrderSpacing leaves room between positions, like StatusOrderSpacing.
	FieldOrderSpacing = 10

	FieldOpEq       = "eq"
	FieldOpAny      = "any"
	FieldOpContains = "contains"
	FieldOpRange    = "range"

	fieldSortPrefix = "field:"
	isoDate         = "2006-01-02"
)

var (
	ErrFieldNotFound          = errors.New("custom field not found")
	ErrFieldNameRequired      = errors.New("field name is required")
	ErrFieldNameTooLong       = errors.New("field name is too long")
	ErrFieldNameTaken         = errors.New("a field with that name already exists in this project")
	ErrInvalidFieldType       = errors.New("unknown field type")
	ErrFieldLimit             = errors.New("project has the maximum number of fields")
	ErrFieldOptionsRequired   = errors.New("a select field needs at least one option")
	ErrFieldOptionsNotAllowed = errors.New("only select fields have options")
	ErrFieldOptionLimit       = errors.New("too many options")
	ErrFieldOptionLabel       = errors.New("option label is empty or too long")
	ErrFieldOptionDuplicate   = errors.New("option labels must be unique")
	ErrFieldOptionUnknown     = errors.New("option id does not belong to this field")
	ErrFieldOrderMismatch     = errors.New("field order must list every field exactly once")
	ErrFieldFilterInvalid     = errors.New("invalid custom field filter")
	ErrFieldSortInvalid       = errors.New("invalid custom field sort")
	ErrTooManyFieldValues     = errors.New("too many field values in one request")
)

// Codes carried by FieldValueError.
const (
	FieldErrUnknownField = "unknown_field"
	FieldErrRequired     = "required"
	FieldErrShape        = "wrong_shape"
	FieldErrTextTooLong  = "text_too_long"
	FieldErrNumber       = "invalid_number"
	FieldErrDate         = "invalid_date"
	FieldErrOption       = "invalid_option"
	FieldErrPerson       = "invalid_person"
	FieldErrURL          = "invalid_url"
	FieldErrCheckbox     = "invalid_checkbox"
	FieldErrTooMany      = "too_many_values"
)

// FieldValueError says which field refused a value and why, so the API can
// answer with a message naming the field.
type FieldValueError struct {
	Code      string
	FieldID   int64
	FieldName string
}

func (e *FieldValueError) Error() string {
	return "custom field " + strconv.FormatInt(e.FieldID, 10) + ": " + e.Code
}

type FieldOption struct {
	ID    string
	Label string
	Color string
}

// FieldValue is the wire and storage-neutral form of a value: Value for the
// single-value types, Values for multi_select and person.
type FieldValue struct {
	Value  string
	Values []string
}

func (v FieldValue) IsEmpty() bool { return v.Value == "" && len(v.Values) == 0 }

type CustomField struct {
	ID         int64
	BusinessID int64
	ProjectID  int64
	Name       string
	Type       string
	Options    []FieldOption
	Required   bool
	Default    FieldValue
	SortOrder  int32
	// ValueCount is how many tasks hold a value. Only ListProjectFields fills it.
	ValueCount int32
	CreatedBy  int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CustomFieldInput struct {
	ProjectID int64
	Name      string
	Type      string
	Options   []FieldOption
	Required  bool
	Default   FieldValue
}

// FieldValueInput is one entry of a batch write. An empty Value clears.
type FieldValueInput struct {
	FieldID int64
	Value   FieldValue
}

// FieldValueWrite is a validated write. An empty Value deletes the row.
type FieldValueWrite struct {
	Field CustomField
	Value FieldValue
}

// TaskFieldFilter is one validated listTasks clause on a custom field.
type TaskFieldFilter struct {
	Field  CustomField
	Op     string
	Values []string
}

type TaskFieldSort struct {
	Field CustomField
	Desc  bool
}

// RawFieldFilter is a clause as the client sends it, before the field is known.
type RawFieldFilter struct {
	FieldID int64
	Op      string
	Values  []string
}

func ValidFieldType(t string) bool {
	switch t {
	case FieldTypeText, FieldTypeNumber, FieldTypeDate, FieldTypeSelect,
		FieldTypeMultiSelect, FieldTypePerson, FieldTypeURL, FieldTypeCheckbox:
		return true
	}
	return false
}

func fieldHasOptions(t string) bool { return t == FieldTypeSelect || t == FieldTypeMultiSelect }

func fieldIsMulti(t string) bool { return t == FieldTypeMultiSelect || t == FieldTypePerson }

func newFieldOptionID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// NormalizeFieldInput validates a create or replace. existing is the field's
// current option list (nil on create). It returns the input with option ids
// assigned and the ids of existing options the new list leaves out.
func NormalizeFieldInput(in CustomFieldInput, existing []FieldOption) (CustomFieldInput, []string, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, nil, ErrFieldNameRequired
	}
	if utf8.RuneCountInString(in.Name) > MaxFieldNameLength {
		return in, nil, ErrFieldNameTooLong
	}
	if !ValidFieldType(in.Type) {
		return in, nil, ErrInvalidFieldType
	}
	options, removed, err := normalizeFieldOptions(in.Type, in.Options, existing)
	if err != nil {
		return in, nil, err
	}
	in.Options = options

	field := CustomField{Name: in.Name, Type: in.Type, Options: options}
	in.Default = resolveOptionLabels(field, in.Default)
	in.Default, err = NormalizeFieldValue(field, in.Default)
	if err != nil {
		return in, nil, err
	}
	return in, removed, nil
}

func normalizeFieldOptions(fieldType string, options, existing []FieldOption) ([]FieldOption, []string, error) {
	if !fieldHasOptions(fieldType) {
		if len(options) > 0 {
			return nil, nil, ErrFieldOptionsNotAllowed
		}
		return []FieldOption{}, nil, nil
	}
	if len(options) == 0 {
		return nil, nil, ErrFieldOptionsRequired
	}
	if len(options) > MaxFieldOptions {
		return nil, nil, ErrFieldOptionLimit
	}
	known := make(map[string]bool, len(existing))
	for _, o := range existing {
		known[o.ID] = true
	}
	out := make([]FieldOption, 0, len(options))
	labels := make(map[string]bool, len(options))
	kept := make(map[string]bool, len(options))
	for _, o := range options {
		o.Label = strings.TrimSpace(o.Label)
		if o.Label == "" || utf8.RuneCountInString(o.Label) > MaxFieldOptionLabel {
			return nil, nil, ErrFieldOptionLabel
		}
		key := strings.ToLower(o.Label)
		if labels[key] {
			return nil, nil, ErrFieldOptionDuplicate
		}
		labels[key] = true
		if o.ID == "" {
			id, err := newFieldOptionID()
			if err != nil {
				return nil, nil, err
			}
			o.ID = id
		} else if !known[o.ID] || kept[o.ID] {
			return nil, nil, ErrFieldOptionUnknown
		}
		kept[o.ID] = true
		if strings.TrimSpace(o.Color) == "" {
			o.Color = DefaultFieldOptionColor
		}
		out = append(out, o)
	}
	var removed []string
	for _, o := range existing {
		if !kept[o.ID] {
			removed = append(removed, o.ID)
		}
	}
	return out, removed, nil
}

// resolveOptionLabels lets a default name an option by label, which is the
// only way to point at an option whose id the server has not assigned yet.
func resolveOptionLabels(field CustomField, v FieldValue) FieldValue {
	if !fieldHasOptions(field.Type) {
		return v
	}
	resolve := func(s string) string {
		for _, o := range field.Options {
			if o.ID == s {
				return s
			}
		}
		for _, o := range field.Options {
			if strings.EqualFold(o.Label, strings.TrimSpace(s)) {
				return o.ID
			}
		}
		return s
	}
	if v.Value != "" {
		v.Value = resolve(v.Value)
	}
	if len(v.Values) > 0 {
		resolved := make([]string, 0, len(v.Values))
		for _, s := range v.Values {
			resolved = append(resolved, resolve(s))
		}
		v.Values = resolved
	}
	return v
}

var fieldNumberPattern = regexp.MustCompile(`^-?\d{1,15}(\.\d{1,6})?$`)

// CanonicalFieldNumber trims a decimal to its shortest form, so "12.50" and
// the "12.500000" MySQL returns both read back as "12.5".
func CanonicalFieldNumber(s string) string {
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	if s == "-0" || s == "" {
		return "0"
	}
	return s
}

// NormalizeFieldValue validates a value against its field and returns the
// canonical form. An empty input returns an empty value and no error: whether
// clearing is allowed is the caller's decision (see Required). It cannot tell
// whether a person may see the project; the service checks that.
func NormalizeFieldValue(field CustomField, in FieldValue) (FieldValue, error) {
	fail := func(code string) (FieldValue, error) {
		return FieldValue{}, &FieldValueError{Code: code, FieldID: field.ID, FieldName: field.Name}
	}
	if fieldIsMulti(field.Type) {
		if in.Value != "" {
			return fail(FieldErrShape)
		}
		return normalizeMultiFieldValue(field, in.Values, fail)
	}
	if len(in.Values) > 0 {
		return fail(FieldErrShape)
	}
	value := strings.TrimSpace(in.Value)
	if value == "" {
		return FieldValue{}, nil
	}
	switch field.Type {
	case FieldTypeText:
		if utf8.RuneCountInString(value) > MaxFieldTextLength {
			return fail(FieldErrTextTooLong)
		}
	case FieldTypeURL:
		u, err := url.Parse(value)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fail(FieldErrURL)
		}
		if utf8.RuneCountInString(value) > MaxFieldTextLength {
			return fail(FieldErrTextTooLong)
		}
	case FieldTypeNumber:
		if !fieldNumberPattern.MatchString(value) {
			return fail(FieldErrNumber)
		}
		value = CanonicalFieldNumber(value)
	case FieldTypeDate:
		if _, err := time.Parse(isoDate, value); err != nil {
			return fail(FieldErrDate)
		}
	case FieldTypeCheckbox:
		if value != "true" && value != "false" {
			return fail(FieldErrCheckbox)
		}
	case FieldTypeSelect:
		if !fieldHasOption(field, value) {
			return fail(FieldErrOption)
		}
	default:
		return fail(FieldErrShape)
	}
	return FieldValue{Value: value}, nil
}

func normalizeMultiFieldValue(field CustomField, values []string, fail func(string) (FieldValue, error)) (FieldValue, error) {
	limit := MaxFieldOptions
	if field.Type == FieldTypePerson {
		limit = MaxFieldPeople
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, raw := range values {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if field.Type == FieldTypePerson {
			id, err := strconv.ParseInt(v, 10, 64)
			if err != nil || id <= 0 {
				return fail(FieldErrPerson)
			}
			v = strconv.FormatInt(id, 10)
		} else if !fieldHasOption(field, v) {
			return fail(FieldErrOption)
		}
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	if len(out) > limit {
		return fail(FieldErrTooMany)
	}
	if len(out) == 0 {
		return FieldValue{}, nil
	}
	return FieldValue{Values: out}, nil
}

func fieldHasOption(field CustomField, id string) bool {
	for _, o := range field.Options {
		if o.ID == id {
			return true
		}
	}
	return false
}

// ValidateFieldOrder reports whether ids is exactly the current set.
func ValidateFieldOrder(current, ids []int64) error {
	if len(ids) != len(current) || len(ids) == 0 {
		return ErrFieldOrderMismatch
	}
	want := make(map[int64]bool, len(current))
	for _, id := range current {
		want[id] = true
	}
	for _, id := range ids {
		if !want[id] {
			return ErrFieldOrderMismatch
		}
		delete(want, id)
	}
	return nil
}

// BuildTaskFieldFilter checks a clause against its field: the operator must
// suit the type and every value must be one the field could hold.
func BuildTaskFieldFilter(field CustomField, op string, values []string) (TaskFieldFilter, error) {
	out := TaskFieldFilter{Field: field, Op: op}
	check := func(v string) (string, error) {
		if field.Type == FieldTypeText || field.Type == FieldTypeURL {
			return v, nil
		}
		in := FieldValue{Value: v}
		if fieldIsMulti(field.Type) {
			in = FieldValue{Values: []string{v}}
		}
		normalized, err := NormalizeFieldValue(field, in)
		if err != nil || normalized.IsEmpty() {
			return "", ErrFieldFilterInvalid
		}
		if fieldIsMulti(field.Type) {
			return normalized.Values[0], nil
		}
		return normalized.Value, nil
	}
	switch op {
	case FieldOpEq:
		if fieldIsMulti(field.Type) || len(values) != 1 {
			return out, ErrFieldFilterInvalid
		}
	case FieldOpAny:
		if (field.Type != FieldTypeSelect && !fieldIsMulti(field.Type)) || len(values) == 0 || len(values) > MaxFieldOptions {
			return out, ErrFieldFilterInvalid
		}
	case FieldOpContains:
		if (field.Type != FieldTypeText && field.Type != FieldTypeURL) || len(values) != 1 || strings.TrimSpace(values[0]) == "" {
			return out, ErrFieldFilterInvalid
		}
	case FieldOpRange:
		if (field.Type != FieldTypeNumber && field.Type != FieldTypeDate) || len(values) != 2 || (values[0] == "" && values[1] == "") {
			return out, ErrFieldFilterInvalid
		}
	default:
		return out, ErrFieldFilterInvalid
	}
	for _, v := range values {
		if op == FieldOpRange && v == "" {
			out.Values = append(out.Values, "")
			continue
		}
		normalized, err := check(v)
		if err != nil {
			return out, err
		}
		out.Values = append(out.Values, normalized)
	}
	return out, nil
}

// ParseFieldSort reads a "field:<id>" sort key. ok is false for any other key,
// which keeps the board order.
func ParseFieldSort(sortBy string) (fieldID int64, ok bool, err error) {
	if !strings.HasPrefix(sortBy, fieldSortPrefix) {
		return 0, false, nil
	}
	id, parseErr := strconv.ParseInt(strings.TrimPrefix(sortBy, fieldSortPrefix), 10, 64)
	if parseErr != nil || id <= 0 {
		return 0, false, ErrFieldSortInvalid
	}
	return id, true, nil
}

// FieldSortable reports whether tasks can be ordered by a field of this type.
func FieldSortable(fieldType string) bool { return !fieldIsMulti(fieldType) }
