package scopie

import (
	"log/slog"
	"strings"
)

const (
	BlockSeperator = "/"
	ScopeSeperator = ","
	ArraySeperator = "|"
	VariablePrefix = "@"
	Wildcard       = "*"
	SuperWildcard  = "**"
)

type Result string

const (
	ResultUnknown Result = "unknown"
	ResultAllow   Result = "allow"
	ResultDeny    Result = "deny"
)

// Logger can be used to override the default logger,
// otherwise it will use slog.Default().
// Scopie only ever logs to debug level.
var Logger *slog.Logger

// Process ...
func Process(vars map[string]string, actorScopes, requiredRules string) (Result, error) {
	logger := Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger = logger.With(
		"actorScopes", actorScopes,
		"requiredRules", requiredRules,
		"variables", vars,
	)

	// do this the simplest way for now, efficiency can come later...
	hasBeenAllowed := false
	logger.Debug("processing scopes")

	actorScopesSplit := strings.Split(actorScopes, ScopeSeperator)
	ruleScopesSplit := strings.Split(requiredRules, ScopeSeperator)
	ruleScopes := make([][]string, len(ruleScopesSplit))
	for i, ruleScope := range ruleScopesSplit {
		ruleScopes[i] = strings.Split(ruleScope, BlockSeperator)
	}

	for _, actorScope := range actorScopesSplit {
		actorSplit := strings.Split(actorScope, BlockSeperator)
		actorScopes := actorSplit[1:]
		rule := actorSplit[0]

		// we can skip this allow rule since we were already approved
		// if hasBeenAllowed && rule == string(ResultAllow) {
		// continue
		// }

		for _, ruleScope := range ruleScopes {
			if isMatch(vars, actorScopes, ruleScope) {
				logger.Debug(
					"matched actor and rule",
					"actorScope", actorScope,
					"rule", ruleScope,
				)

				if rule == string(ResultDeny) {
					logger.Debug(
						"matched deny rule",
						"actorScope", actorScope,
						"rule", ruleScope,
					)

					return ResultDeny, nil
				} else {
					logger.Debug(
						"matched allow rule",
						"actorScope", actorScope,
						"rule", ruleScope,
					)

					hasBeenAllowed = true
				}
			}
		}
	}

	if hasBeenAllowed {
		logger.Debug("has been allowed")
		return ResultAllow, nil
	}

	// return unknown until we are done testing our cases for now
	// otherwise we should return unknown for errors at least
	return ResultUnknown, nil
}

func isMatch(vars map[string]string, actorScope, ruleScope []string) bool {
	slog.Info("checking a match", "actor", actorScope, "rule", ruleScope)
NextRule:
	for i, ruleBlock := range ruleScope {
        if len(actorScope) <= i {
            return false
        }
		actorBlock := actorScope[i]
		if actorBlock == Wildcard {
			continue
		}

		if actorBlock == SuperWildcard {
			return true
		}

		if strings.Contains(actorBlock, ArraySeperator) {
			slog.Info("comparing rule array to scope", "rule array", ruleBlock)
			for _, actorArrayValue := range strings.Split(actorBlock, ArraySeperator) {
				if actorArrayValue == ruleBlock {
					slog.Info("found matching array value", "value", actorArrayValue)
					continue NextRule
				}
			}
			return false
		} else if strings.HasPrefix(actorBlock, VariablePrefix) {
			if vars[strings.TrimPrefix(actorBlock, VariablePrefix)] != ruleBlock {
				return false
			}
		} else if ruleBlock != actorBlock {
			return false
		}
	}

	return true
}
