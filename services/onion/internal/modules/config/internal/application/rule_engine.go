package application

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
)

func (s *Service) EvaluateRule(userContext *domain.FeatureContext, rule *domain.FeatureRule) bool {
	if userContext == nil {
		return false
	}

	for key, value := range rule.Conditions {
		match := s.evaluateCondition(userContext, key, value)
		if !match {
			return false
		}
	}

	return true
}

func (s *Service) evaluateCondition(userContext *domain.FeatureContext, key string, value interface{}) bool {
	switch key {
	case "user_id":
		return s.evaluateUserID(userContext, value)
	case "user_type":
		return s.evaluateUserType(userContext, value)
	case "user_role":
		return s.evaluateUserRole(userContext, value)
	case "country":
		return s.evaluateCountry(userContext, value)
	case "platform":
		return s.evaluatePlatform(userContext, value)
	case "app_version":
		return s.evaluateAppVersion(userContext, value)
	case "registered_at":
		return s.evaluateRegisteredAt(userContext, value)
	case "is_anonymous":
		return s.evaluateIsAnonymous(userContext, value)
	default:
		return s.evaluateCustomAttribute(userContext, key, value)
	}
}

func (s *Service) evaluateUserID(userContext *domain.FeatureContext, value interface{}) bool {
	switch v := value.(type) {
	case float64:
		return userContext.UserID == int32(v)
	case []interface{}:
		for _, id := range v {
			if idFloat, ok := id.(float64); ok && userContext.UserID == int32(idFloat) {
				return true
			}
		}

		return false
	default:
		return false
	}
}

func (s *Service) evaluateUserType(userContext *domain.FeatureContext, value interface{}) bool {
	return s.evaluateStringCondition(userContext.UserType, value)
}

func (s *Service) evaluateUserRole(userContext *domain.FeatureContext, value interface{}) bool {
	return s.evaluateStringCondition(userContext.UserRole, value)
}

func (s *Service) evaluateCountry(userContext *domain.FeatureContext, value interface{}) bool {
	return s.evaluateStringCondition(userContext.Country, value)
}

func (s *Service) evaluatePlatform(userContext *domain.FeatureContext, value interface{}) bool {
	return s.evaluateStringCondition(userContext.Platform, value)
}

func (s *Service) evaluateAppVersion(userContext *domain.FeatureContext, value interface{}) bool {
	valueStr, ok := value.(string)
	if !ok {
		return false
	}

	if strings.HasPrefix(valueStr, ">=") {
		targetVersion := strings.TrimPrefix(valueStr, ">=")

		return s.compareVersions(userContext.AppVersion, targetVersion) >= 0
	}

	if strings.HasPrefix(valueStr, "<=") {
		targetVersion := strings.TrimPrefix(valueStr, "<=")

		return s.compareVersions(userContext.AppVersion, targetVersion) <= 0
	}

	if strings.HasPrefix(valueStr, ">") {
		targetVersion := strings.TrimPrefix(valueStr, ">")

		return s.compareVersions(userContext.AppVersion, targetVersion) > 0
	}

	if strings.HasPrefix(valueStr, "<") {
		targetVersion := strings.TrimPrefix(valueStr, "<")

		return s.compareVersions(userContext.AppVersion, targetVersion) < 0
	}

	if strings.HasPrefix(valueStr, "=") {
		targetVersion := strings.TrimPrefix(valueStr, "=")

		return userContext.AppVersion == targetVersion
	}

	return userContext.AppVersion == valueStr
}

func (s *Service) evaluateRegisteredAt(userContext *domain.FeatureContext, value interface{}) bool {
	if userContext.RegisteredAt == nil {
		return false
	}

	valueStr, ok := value.(string)
	if !ok {
		return false
	}

	if strings.HasSuffix(valueStr, "_days_ago") {
		daysStr := strings.TrimSuffix(valueStr, "_days_ago")

		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return false
		}

		cutoff := time.Now().AddDate(0, 0, -days)

		return userContext.RegisteredAt.Before(cutoff)
	}

	targetTime, err := time.Parse(time.RFC3339, valueStr)
	if err != nil {
		return false
	}

	return userContext.RegisteredAt.Before(targetTime)
}

func (s *Service) evaluateIsAnonymous(userContext *domain.FeatureContext, value interface{}) bool {
	boolValue, ok := value.(bool)
	if !ok {
		return false
	}

	return userContext.IsAnonymous == boolValue
}

func (s *Service) evaluateCustomAttribute(userContext *domain.FeatureContext, key string, value interface{}) bool {
	attrValue := userContext.GetAttribute(key)
	if attrValue == nil {
		return false
	}

	switch v := value.(type) {
	case string:
		if attrStr, ok := attrValue.(string); ok {
			return attrStr == v
		}
	case float64:
		if attrFloat, ok := attrValue.(float64); ok {
			return attrFloat == v
		}

		if attrInt, ok := attrValue.(int); ok {
			return float64(attrInt) == v
		}
	case bool:
		if attrBool, ok := attrValue.(bool); ok {
			return attrBool == v
		}
	case []interface{}:
		return s.evaluateStringCondition(fmt.Sprintf("%v", attrValue), value)
	}

	return false
}

func (s *Service) evaluateStringCondition(userValue string, value interface{}) bool {
	switch v := value.(type) {
	case string:
		return userValue == v
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok && userValue == str {
				return true
			}
		}

		return false
	default:
		return false
	}
}

func (s *Service) compareVersions(version1, version2 string) int {
	v1Parts := strings.Split(version1, ".")
	v2Parts := strings.Split(version2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1Part, v2Part int

		if i < len(v1Parts) {
			if val, err := strconv.Atoi(v1Parts[i]); err == nil {
				v1Part = val
			}
		}

		if i < len(v2Parts) {
			if val, err := strconv.Atoi(v2Parts[i]); err == nil {
				v2Part = val
			}
		}

		if v1Part < v2Part {
			return -1
		}

		if v1Part > v2Part {
			return 1
		}
	}

	return 0
}
