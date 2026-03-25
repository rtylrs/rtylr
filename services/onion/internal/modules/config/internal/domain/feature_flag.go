package domain

import (
	"encoding/json"
	"time"

	sqlc "github.com/voxire/onion/internal/modules/config/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

type FeatureFlag struct {
	Id                int32
	Key               string
	Name              string
	Description       string
	Enabled           bool
	RolloutPercentage int32
}

type FeatureRule struct {
	Id                int32
	Name              string
	Conditions        map[string]interface{}
	Action            string
	RolloutPercentage int32
	Priority          int32
	Enabled           bool
}

func (f *FeatureFlag) FromRows(row sqlc.GetFeatureFlagRow) {
	f.Id = row.ID
	f.Key = row.Key
	f.Name = row.Name
	f.Description = convert.ToString(row.Description)
	f.Enabled = convert.ToBool(row.Enabled)
	f.RolloutPercentage = convert.ToInt32(row.RolloutPercentage)
}

type FeatureFlagWithRules struct {
	FeatureFlag FeatureFlag
	Rules       []FeatureRule
}

type FeatureFlagWithRulesRow struct {
	FeatureFlagId            int32
	FeatureKey               string
	FeatureName              string
	FeatureDescription       string
	FeatureEnabled           bool
	FeatureRolloutPercentage int32
	RuleId                   *int32
	RuleName                 string
	RuleConditions           map[string]interface{}
	RuleAction               string
	RuleRolloutPercentage    int32
	RulePriority             int32
	RuleEnabled              bool
}

func (r *FeatureFlagWithRulesRow) FromRows(row sqlc.GetFeatureFlagsWithRulesRow) {
	r.FeatureFlagId = row.FeatureFlagID
	r.FeatureKey = row.FeatureKey
	r.FeatureName = row.FeatureName
	r.FeatureDescription = convert.ToString(row.FeatureDescription)
	r.FeatureEnabled = convert.ToBool(row.FeatureEnabled)
	r.FeatureRolloutPercentage = convert.ToInt32(row.FeatureRolloutPercentage)

	if row.RuleID.Valid {
		r.RuleId = &row.RuleID.Int32
		r.RuleName = convert.ToString(row.RuleName)
		r.parseConditions(row.RuleConditions)
		r.RuleAction = convert.ToString(row.RuleAction)
		r.RuleRolloutPercentage = convert.ToInt32(row.RuleRolloutPercentage)
		r.RulePriority = convert.ToInt32(row.RulePriority)
		r.RuleEnabled = convert.ToBool(row.RuleEnabled)
	}
}

func (r *FeatureFlagWithRulesRow) parseConditions(conditionsBytes []byte) {
	if len(conditionsBytes) == 0 {
		r.RuleConditions = make(map[string]interface{})

		return
	}

	var conditions map[string]interface{}
	if err := json.Unmarshal(conditionsBytes, &conditions); err != nil {
		r.RuleConditions = make(map[string]interface{})

		return
	}

	r.RuleConditions = conditions
}

func GroupFeatureFlagsWithRules(rows []FeatureFlagWithRulesRow) []FeatureFlagWithRules {
	flagMap := make(map[string]*FeatureFlagWithRules)

	for i := range rows {
		row := rows[i]
		key := row.FeatureKey

		if _, exists := flagMap[key]; !exists {
			flagMap[key] = &FeatureFlagWithRules{
				FeatureFlag: FeatureFlag{
					Id:                row.FeatureFlagId,
					Key:               row.FeatureKey,
					Name:              row.FeatureName,
					Description:       row.FeatureDescription,
					Enabled:           row.FeatureEnabled,
					RolloutPercentage: row.FeatureRolloutPercentage,
				},
				Rules: []FeatureRule{},
			}
		}

		if row.RuleId != nil {
			rule := FeatureRule{
				Id:                *row.RuleId,
				Name:              row.RuleName,
				Conditions:        row.RuleConditions,
				Action:            row.RuleAction,
				RolloutPercentage: row.RuleRolloutPercentage,
				Priority:          row.RulePriority,
				Enabled:           row.RuleEnabled,
			}
			flagMap[key].Rules = append(flagMap[key].Rules, rule)
		}
	}

	result := make([]FeatureFlagWithRules, 0, len(flagMap))
	for _, flagWithRules := range flagMap {
		result = append(result, *flagWithRules)
	}

	return result
}

type FeatureContext struct {
	UserID       int32
	UserType     string
	UserRole     string
	Country      string
	Platform     string
	AppVersion   string
	RegisteredAt *time.Time
	IsAnonymous  bool
	CustomAttrs  map[string]interface{}
	IPAddress    string
	UserAgent    string
	SessionID    string
}

type featureContextJSON struct {
	UserID       int32                  `json:"user_id"`
	UserType     string                 `json:"user_type"`
	UserRole     string                 `json:"user_role"`
	Country      string                 `json:"country"`
	Platform     string                 `json:"platform"`
	AppVersion   string                 `json:"app_version"`
	RegisteredAt *time.Time             `json:"registered_at"`
	IsAnonymous  bool                   `json:"is_anonymous"`
	CustomAttrs  map[string]interface{} `json:"custom_attrs"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	SessionID    string                 `json:"session_id"`
}

func (c *FeatureContext) ToJSON() ([]byte, error) {
	jsonStruct := featureContextJSON{
		UserID:       c.UserID,
		UserType:     c.UserType,
		UserRole:     c.UserRole,
		Country:      c.Country,
		Platform:     c.Platform,
		AppVersion:   c.AppVersion,
		RegisteredAt: c.RegisteredAt,
		IsAnonymous:  c.IsAnonymous,
		CustomAttrs:  c.CustomAttrs,
		IPAddress:    c.IPAddress,
		UserAgent:    c.UserAgent,
		SessionID:    c.SessionID,
	}

	return json.Marshal(jsonStruct)
}

func (c *FeatureContext) FromJSON(data []byte) error {
	var jsonStruct featureContextJSON
	if err := json.Unmarshal(data, &jsonStruct); err != nil {
		return err
	}

	c.UserID = jsonStruct.UserID
	c.UserType = jsonStruct.UserType
	c.UserRole = jsonStruct.UserRole
	c.Country = jsonStruct.Country
	c.Platform = jsonStruct.Platform
	c.AppVersion = jsonStruct.AppVersion
	c.RegisteredAt = jsonStruct.RegisteredAt
	c.IsAnonymous = jsonStruct.IsAnonymous
	c.CustomAttrs = jsonStruct.CustomAttrs
	c.IPAddress = jsonStruct.IPAddress
	c.UserAgent = jsonStruct.UserAgent
	c.SessionID = jsonStruct.SessionID

	return nil
}

func (c *FeatureContext) GetAttribute(key string) interface{} {
	if c.CustomAttrs == nil {
		return nil
	}

	return c.CustomAttrs[key]
}

func (c *FeatureContext) SetAttribute(key string, value interface{}) {
	if c.CustomAttrs == nil {
		c.CustomAttrs = make(map[string]interface{})
	}

	c.CustomAttrs[key] = value
}

type EvaluationResult struct {
	FeatureKey string
	Enabled    bool
	Reason     string
	RuleID     *int32
	Context    *FeatureContext
}

type FeatureEvaluation struct {
	UserID      *int32
	FeatureKey  string
	Enabled     bool
	Reason      string
	RuleID      *int32
	UserContext *FeatureContext
	EvaluatedAt time.Time
}

type UserFeatureOverride struct {
	Id         int32
	UserId     int32
	FeatureKey string
	Enabled    bool
	Reason     string
}

func (o *UserFeatureOverride) FromRows(row sqlc.GetUserFeatureOverridesRow) {
	o.Id = row.ID
	o.UserId = row.UserID
	o.FeatureKey = row.FeatureKey
	o.Enabled = row.Enabled
	o.Reason = convert.ToString(row.Reason)
}

type Features struct {
	Flags       map[string]bool
	IsAnonymous bool
	SessionId   string
}
