package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

// leadQuery holds the lead filter parameters. The stats request carries
// exactly these; the list request carries them among its own.
type leadQuery crmsalesv1.GetLeadStatsRequest

func listQuery(req *crmsalesv1.ListLeadsRequest) leadQuery {
	return leadQuery{
		Reps: req.Reps, Sources: req.Sources, Industries: req.Industries, Sizes: req.Sizes,
		Countries: req.Countries, Cities: req.Cities, Billing: req.Billing, ImportId: req.ImportId,
		FollowUp: req.FollowUp, HasPhone: req.HasPhone, HasEmail: req.HasEmail, HasWebsite: req.HasWebsite,
		HasLinkedin: req.HasLinkedin, HasCompany: req.HasCompany, HasValue: req.HasValue, Contacted: req.Contacted,
		AttemptsMin: req.AttemptsMin, AttemptsMax: req.AttemptsMax, ValueMin: req.ValueMin, ValueMax: req.ValueMax,
		CreatedFrom: req.CreatedFrom, CreatedTo: req.CreatedTo, ActiveFrom: req.ActiveFrom, ActiveTo: req.ActiveTo,
		Search: req.Search, Today: req.Today, UtcOffsetMinutes: req.UtcOffsetMinutes,
	}
}

// filter parses the parameters; the service checks the values.
func (q leadQuery) filter() (domain.LeadFilter, error) {
	f := domain.LeadFilter{
		ImportBatchID: positivePtr(q.ImportId),
		FollowUp:      domain.FollowUpWindow(q.FollowUp),
		HasPhone:      domain.Presence(q.HasPhone),
		HasEmail:      domain.Presence(q.HasEmail),
		HasWebsite:    domain.Presence(q.HasWebsite),
		HasLinkedIn:   domain.Presence(q.HasLinkedin),
		HasCompany:    domain.Presence(q.HasCompany),
		HasValue:      domain.Presence(q.HasValue),
		Contacted:     domain.Presence(q.Contacted),
		AttemptsMin:   q.AttemptsMin,
		AttemptsMax:   q.AttemptsMax,
		Search:        q.Search,
	}
	for _, rep := range splitList(q.Reps) {
		id, err := strconv.ParseInt(rep, 10, 64)
		if err != nil || id < 0 {
			return f, invalid("reps must be user ids")
		}
		f.Reps = append(f.Reps, id)
	}
	for _, billing := range splitList(q.Billing) {
		f.Billing = append(f.Billing, domain.Billing(billing))
	}
	var err error
	for _, set := range []struct {
		value, field string
		into         *[]string
	}{
		{q.Sources, "sources", &f.Sources},
		{q.Industries, "industries", &f.Industries},
		{q.Sizes, "sizes", &f.Sizes},
		{q.Countries, "countries", &f.Countries},
		{q.Cities, "cities", &f.Cities},
	} {
		if *set.into, err = jsonList(set.value, set.field); err != nil {
			return f, err
		}
	}
	if f.ValueMin, err = parseAmountPtr(&q.ValueMin, "value_min"); err != nil {
		return f, err
	}
	if f.ValueMax, err = parseAmountPtr(&q.ValueMax, "value_max"); err != nil {
		return f, err
	}
	for _, bound := range []struct {
		value, field string
		into         **time.Time
	}{
		{q.CreatedFrom, "created_from", &f.CreatedFrom},
		{q.CreatedTo, "created_to", &f.CreatedTo},
		{q.ActiveFrom, "active_from", &f.ActiveFrom},
		{q.ActiveTo, "active_to", &f.ActiveTo},
	} {
		if *bound.into, err = parseDatePtr(bound.value, bound.field); err != nil {
			return f, err
		}
	}
	return f, nil
}

// jsonList reads a query parameter holding a JSON array of strings, which
// keeps values with commas intact.
func jsonList(value, field string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var out []string
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil, invalid(field + " must be a JSON array of strings")
	}
	return out, nil
}

func parseDatePtr(value, field string) (*time.Time, error) {
	if value = strings.TrimSpace(value); value == "" {
		return nil, nil
	}
	day, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, invalid(field + " must be a date (YYYY-MM-DD)")
	}
	return &day, nil
}
