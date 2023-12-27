package scopie

import (
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

// Process ...
func Process(vars map[string]string, actorScopes, requiredRules string) (Result, error) {
	// do this the simplest way for now, efficiency can come later...
	hasBeenAllowed := false

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
		if hasBeenAllowed && rule == string(ResultAllow) {
			continue
		}

		for _, ruleScope := range ruleScopes {
			if isMatch(vars, actorScopes, ruleScope) {
				if rule == string(ResultDeny) {
					return ResultDeny, nil
				} else {
					hasBeenAllowed = true
				}
			}
		}
	}

	if hasBeenAllowed {
		return ResultAllow, nil
	}

	// return unknown until we are done testing our cases for now
	// otherwise we should return unknown for errors at least
	return ResultUnknown, nil
}

func isMatch(vars map[string]string, actorScope, ruleScope []string) bool {
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
			for _, actorArrayValue := range strings.Split(actorBlock, ArraySeperator) {
				if actorArrayValue == ruleBlock {
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
