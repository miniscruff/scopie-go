package scopie

import (
	"log/slog"
	"slices"
	"strings"
)

const (
	BlockSeperator = "/"
	ScopeSeperator = ","
	ArraySeperator = "|"
)

type Result string

const (
	ResultUnknown Result = "uknown"
	ResultAllow   Result = "allow"
	ResultDeny    Result = "deny"
)

// Logger can be used to override the default logger,
// otherwise it will use slog.Default().
// Scopie only ever logs to debug level.
var Logger *slog.Logger

// Process ...
func Process(actorScopes, requiredRules string) (Result, error) {
	logger := Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger = logger.With(
		"actorScopes", actorScopes,
		"requiredRules", requiredRules,
	)

	// do this the simplest way for now, efficiency can come later...
	hasBeenAllowed := false
	logger.Debug("processing scopes")

	actorScopesSplit := strings.Split(actorScopes, ScopeSeperator)
	actorScopesSplitExpanded := make([]string, 0)
	for _, actorScope := range actorScopesSplit {
		actorScopesSplitExpanded = append(actorScopesSplitExpanded, expandVars(actorScope)...)
	}

	logger.Debug("expanded scopes", "expanded", actorScopesSplitExpanded)

	for _, actorScope := range actorScopesSplitExpanded {
		actorSplit := strings.Split(actorScope, BlockSeperator)
		actorScopes := actorSplit[1:]
		rule := actorSplit[0]

		// we can skip this allow rule since we were already approved
		if hasBeenAllowed && rule == string(ResultAllow) {
			continue
		}

		for _, requiredRule := range strings.Split(requiredRules, ScopeSeperator) {
			ruleScopes := strings.Split(requiredRule, BlockSeperator)

			if slices.Equal(ruleScopes, actorScopes) {
				logger.Debug(
					"matched actor and rule",
					"actorScope", actorScope,
					"rule", requiredRule,
				)

				if rule == string(ResultDeny) {
					logger.Debug(
						"matched deny rule",
						"actorScope", actorScope,
						"rule", requiredRule,
					)

					return ResultDeny, nil
				} else {
					logger.Debug(
						"matched allow rule",
						"actorScope", actorScope,
						"rule", requiredRule,
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

// expandVars takes a string that contains a/[b,c,d] lists and expands to a/b,a/c,a/d
func expandVars(value string) []string {
	if !strings.Contains(value, ArraySeperator) {
		return []string{value}
	}

	ret := make([]string, 0)
	blocks := strings.Split(value, BlockSeperator)
	blocksCopy := make([]string, len(blocks))
	for i, b := range blocks {
		if strings.Contains(b, ArraySeperator) {
			copy(blocksCopy, blocks)
			for _, arrayValue := range strings.Split(b, ArraySeperator) {
				blocksCopy[i] = arrayValue
				ret = append(ret, expandVars(strings.Join(blocksCopy, BlockSeperator))...)
			}
		}
	}

	return ret
}
