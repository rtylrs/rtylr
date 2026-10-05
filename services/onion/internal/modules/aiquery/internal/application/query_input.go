package application

import "github.com/voxire/onion/internal/modules/aiquery/internal/domain"

type queryInput struct {
	View         string   `json:"view"`
	Columns      []string `json:"columns"`
	Aggregations []struct {
		Func string `json:"func"`
		Col  string `json:"col"`
		As   string `json:"as"`
	} `json:"aggregations"`
	Where []struct {
		Col         string `json:"col"`
		Op          string `json:"op"`
		Value       any    `json:"value"`
		ValueColumn string `json:"value_col"`
	} `json:"where"`
	GroupBy []string `json:"group_by"`
	OrderBy []struct {
		Col string `json:"col"`
		Dir string `json:"dir"`
	} `json:"order_by"`
	Limit int `json:"limit"`
}

func (input queryInput) spec() domain.QuerySpec {
	spec := domain.QuerySpec{View: input.View, Columns: input.Columns, GroupBy: input.GroupBy, Limit: input.Limit}
	for _, aggregation := range input.Aggregations {
		spec.Aggregations = append(spec.Aggregations, domain.Aggregation{Func: aggregation.Func, Col: aggregation.Col, As: aggregation.As})
	}
	for _, filter := range input.Where {
		spec.Where = append(spec.Where, domain.Filter{Col: filter.Col, Op: filter.Op, Value: filter.Value, ValueColumn: filter.ValueColumn})
	}
	for _, order := range input.OrderBy {
		spec.OrderBy = append(spec.OrderBy, domain.OrderBy{Col: order.Col, Dir: order.Dir})
	}
	return spec
}
