package validator

import (
	"fmt"

	"github.com/quantmind-br/llama-cpp-loader/internal/domain"
)

func applyTypeRules(p domain.Profile, schema domain.FlagSchema, rep Report) Report {
	for key, val := range p.Args {
		spec, ok := schema.Lookup(key)
		if !ok {
			continue // unknown flags are not type-checked here
		}
		if msg := checkType(spec, val); msg != "" {
			rep = appendIssue(rep, FieldIssue{Field: key, Message: msg, Severity: SeverityError})
		}
	}
	return rep
}

func checkType(spec domain.FlagSpec, val any) string {
	switch spec.Type {
	case domain.FlagTypeInt:
		switch val.(type) {
		case int, int32, int64, float64, float32:
			return ""
		}
		return fmt.Sprintf("expected int, got %T", val)
	case domain.FlagTypeFloat:
		switch val.(type) {
		case float32, float64, int, int32, int64:
			return ""
		}
		return fmt.Sprintf("expected float, got %T", val)
	case domain.FlagTypeBool:
		if _, ok := val.(bool); ok {
			return ""
		}
		return fmt.Sprintf("expected bool, got %T", val)
	case domain.FlagTypeString:
		if _, ok := val.(string); ok {
			return ""
		}
		return fmt.Sprintf("expected string, got %T", val)
	case domain.FlagTypeEnum:
		s, ok := val.(string)
		if !ok {
			return fmt.Sprintf("expected one of %v, got %T", spec.EnumValues, val)
		}
		for _, v := range spec.EnumValues {
			if v == s {
				return ""
			}
		}
		return fmt.Sprintf("%q not in %v", s, spec.EnumValues)
	}
	return ""
}

func applyCrossFieldRules(p domain.Profile, rep Report) Report {
	return rep
}

func applyExistenceRules(p domain.Profile, rep Report) Report {
	return rep
}
