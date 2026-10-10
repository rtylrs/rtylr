package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
)

func init() {
	registerGeneralTool("utilities.business_math", (*AssistantToolExecutor).businessMath)
}

// businessMath covers the formulas owners ask about most, where a model doing
// the algebra itself tends to slip: margins, VAT, loans, growth, break-even.
func (e *AssistantToolExecutor) businessMath(_ context.Context, input map[string]any) (map[string]any, error) {
	op := strings.ToLower(strings.ReplaceAll(inputText(input, "operation", "op", "formula"), " ", "_"))
	num := func(keys ...string) (float64, error) {
		value, ok := inputNumber(input, keys...)
		if !ok {
			return 0, fmt.Errorf("%s needs %q", op, keys[0])
		}
		return value, nil
	}
	switch op {
	case "margin", "markup":
		cost, err := num("cost")
		if err != nil {
			return nil, err
		}
		price, err := num("price", "sale_price")
		if err != nil {
			return nil, err
		}
		if price == 0 || cost == 0 {
			return nil, errors.New("cost and price must be non-zero")
		}
		return map[string]any{"cost": cost, "price": price, "profit": round2(price - cost),
			"margin_percent": round2((price - cost) / price * 100), "markup_percent": round2((price - cost) / cost * 100)}, nil
	case "price_for_margin":
		cost, err := num("cost")
		if err != nil {
			return nil, err
		}
		margin, err := num("margin_percent", "margin")
		if err != nil {
			return nil, err
		}
		if margin >= 100 {
			return nil, errors.New("margin must be under 100%")
		}
		price := cost / (1 - margin/100)
		return map[string]any{"cost": cost, "margin_percent": margin, "price": round2(price), "profit": round2(price - cost)}, nil
	case "vat", "tax":
		amount, err := num("amount")
		if err != nil {
			return nil, err
		}
		rate, err := num("rate_percent", "rate")
		if err != nil {
			return nil, err
		}
		if inputBool(input, "inclusive") {
			net := amount / (1 + rate/100)
			return map[string]any{"gross": amount, "net": round2(net), "tax": round2(amount - net), "rate_percent": rate}, nil
		}
		return map[string]any{"net": amount, "tax": round2(amount * rate / 100), "gross": round2(amount * (1 + rate/100)), "rate_percent": rate}, nil
	case "percent_change", "growth":
		from, err := num("from", "old", "before")
		if err != nil {
			return nil, err
		}
		to, err := num("to", "new", "after")
		if err != nil {
			return nil, err
		}
		if from == 0 {
			return nil, errors.New("percent change from zero is undefined")
		}
		return map[string]any{"from": from, "to": to, "change": round2(to - from), "percent_change": round2((to - from) / math.Abs(from) * 100)}, nil
	case "cagr":
		start, err := num("start", "from")
		if err != nil {
			return nil, err
		}
		end, err := num("end", "to")
		if err != nil {
			return nil, err
		}
		years, err := num("years", "periods")
		if err != nil {
			return nil, err
		}
		if start <= 0 || end <= 0 || years <= 0 {
			return nil, errors.New("start, end and years must be positive")
		}
		return map[string]any{"cagr_percent": round2((math.Pow(end/start, 1/years) - 1) * 100)}, nil
	case "loan_payment", "loan":
		principal, err := num("principal", "amount")
		if err != nil {
			return nil, err
		}
		annual, err := num("annual_rate_percent", "rate")
		if err != nil {
			return nil, err
		}
		months, err := num("months", "term_months")
		if err != nil {
			return nil, err
		}
		if months <= 0 || months > 1200 {
			return nil, errors.New("months must be between 1 and 1200")
		}
		r := annual / 100 / 12
		payment := principal / months
		if r != 0 {
			payment = principal * r / (1 - math.Pow(1+r, -months))
		}
		total := payment * months
		return map[string]any{"monthly_payment": round2(payment), "total_paid": round2(total), "total_interest": round2(total - principal)}, nil
	case "compound_interest", "compound", "savings":
		principal, err := num("principal", "amount")
		if err != nil {
			return nil, err
		}
		annual, err := num("annual_rate_percent", "rate")
		if err != nil {
			return nil, err
		}
		years, err := num("years")
		if err != nil {
			return nil, err
		}
		perYear, ok := inputNumber(input, "compounds_per_year")
		if !ok || perYear <= 0 {
			perYear = 12
		}
		contribution, _ := inputNumber(input, "monthly_contribution", "contribution")
		months := int(math.Round(years * 12))
		if months <= 0 || months > 1200 {
			return nil, errors.New("years must be between 1 month and 100 years")
		}
		balance, contributed := principal, principal
		monthlyGrowth := math.Pow(1+annual/100/perYear, perYear/12)
		for i := 0; i < months; i++ {
			balance = balance*monthlyGrowth + contribution
			contributed += contribution
		}
		return map[string]any{"final_balance": round2(balance), "total_contributed": round2(contributed), "interest_earned": round2(balance - contributed)}, nil
	case "break_even":
		fixed, err := num("fixed_costs", "fixed")
		if err != nil {
			return nil, err
		}
		price, err := num("price", "unit_price")
		if err != nil {
			return nil, err
		}
		variable, err := num("variable_cost", "unit_cost")
		if err != nil {
			return nil, err
		}
		if price <= variable {
			return nil, errors.New("price must exceed variable cost to break even")
		}
		units := fixed / (price - variable)
		return map[string]any{"break_even_units": math.Ceil(units), "break_even_revenue": round2(math.Ceil(units) * price), "contribution_per_unit": round2(price - variable)}, nil
	case "discount":
		price, err := num("price", "amount")
		if err != nil {
			return nil, err
		}
		discount, err := num("discount_percent", "percent")
		if err != nil {
			return nil, err
		}
		return map[string]any{"original": price, "discount": round2(price * discount / 100), "final": round2(price * (1 - discount/100))}, nil
	default:
		return nil, errors.New("operation must be one of margin, price_for_margin, vat, percent_change, cagr, loan_payment, compound_interest, break_even, discount")
	}
}

func round2(value float64) float64 { return roundTo(value, 2) }
