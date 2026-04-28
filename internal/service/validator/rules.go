package validator

import "github.com/quantmind-br/llama-cpp-loader/internal/domain"

func applyTypeRules(p domain.Profile, schema domain.FlagSchema, rep Report) Report {
	return rep
}

func applyCrossFieldRules(p domain.Profile, rep Report) Report {
	return rep
}

func applyExistenceRules(p domain.Profile, rep Report) Report {
	return rep
}
