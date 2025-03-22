package scopie

import (
	"errors"
	"fmt"
	"strings"
)

const (
	BlockSeperator = byte('/')
	ArraySeperator = byte('|')
	VariablePrefix = byte('@')
	Wildcard       = byte('*')

	AllowPermission = "allow"
	DenyPermission  = "deny"
)

const (
	fmtAllowedInvalidChar = "scopie-100 in %s: invalid character '%s'"
	fmtAllowedVarInArray  = "scopie-101: variable '%s' found in array block"
	fmtAllowedVarNotFound = "scopie-104: variable '%s' not found"

	fmtValidateVarInArray  = "scopie-101: variable '%s' found in array block"
	fmtValidateInvalidChar = "scopie-100: invalid character '%s'"
)

var (
	errSuperNotLast    = errors.New("scopie-105: super wildcard not in the last block")
	errSuperInArray    = errors.New("scopie-103: super wildcard found in array block")
	errWildcardInArray = errors.New("scopie-102: wildcard found in array block")
	errScopesEmpty     = errors.New("scopie-106 in scope: scopes was empty")
	errScopeEmpty      = errors.New("scopie-106 in scope: scope was empty")
	errRuleEmpty       = errors.New("scopie-106 in rule: rule was empty")

	// validation specific
	errValidateScopeRulesEmpty = errors.New("scopie-106: scope or rule was empty")
	errValidateNoScopeRules    = errors.New("scopie-106: scope or rule array was empty")
	errValidateInconsistent    = errors.New("scopie-107: inconsistent array of scopes and rules")
)

// IsAllowedFunc is a type wrapper for IsAllowed that can be used as
// a dependency.
type IsAllowedFunc func(map[string]string, string, string) (bool, error)

// ValidateScopeFunc is a type wrapper for ValidateScope that can be
// used as a dependency.
type ValidateScopeFunc func(string) error

// IsAllowed returns whether or not the required role scopes are fulfilled by our actor scopes.
func IsAllowed(scopes, rules []string, vars map[string]string) (bool, error) {
	if len(scopes) == 0 {
		return false, errScopesEmpty
	}

	if len(rules) == 0 {
		return false, nil
	}

	hasBeenAllowed := false

	for _, actorRule := range rules {
		if len(actorRule) == 0 {
			return false, errRuleEmpty
		}

		actorRule := actorRule

		isAllowBlock := strings.HasPrefix(actorRule, AllowPermission)
		if isAllowBlock && hasBeenAllowed {
			continue
		}

		for _, actionScope := range scopes {
			if len(actionScope) == 0 {
				return false, errScopeEmpty
			}

			actionScope := actionScope

			match, err := compareRuleToScope(&actorRule, &actionScope, vars)
			if err != nil {
				return false, err
			}

			if match && isAllowBlock {
				hasBeenAllowed = true
			} else if match && !isAllowBlock {
				return false, nil
			}
		}
	}

	return hasBeenAllowed, nil
}

// ValidateScopes returns an error if the scope or rules are invalid.
func ValidateScopes(scopeOrRules []string) error {
	if len(scopeOrRules) == 0 {
		return errValidateNoScopeRules
	}

	isRules := strings.HasPrefix(scopeOrRules[0], AllowPermission) ||
		strings.HasPrefix(scopeOrRules[0], DenyPermission)

	for _, scope := range scopeOrRules {
		if scope == "" {
			return errValidateScopeRulesEmpty
		}

		scopeIsRule := strings.HasPrefix(scope, AllowPermission) ||
			strings.HasPrefix(scope, DenyPermission)

		if isRules != scopeIsRule {
			return errValidateInconsistent
		}

		inArray := false

		for i := range scope {
			if scope[i] == BlockSeperator {
				inArray = false
				continue
			}

			if scope[i] == ArraySeperator {
				inArray = true
				continue
			}

			if inArray {
				if scope[i] == Wildcard && i < len(scope)-1 && scope[i+1] == Wildcard {
					return errSuperInArray
				}

				if scope[i] == Wildcard {
					return errWildcardInArray
				}

				if scope[i] == VariablePrefix {
					end := endOfArrayElement(&scope, i)
					return fmt.Errorf(fmtValidateVarInArray, scope[i+1:end])
				}
			}

			if !isValidCharacter(scope[i]) {
				return fmt.Errorf(fmtValidateInvalidChar, string(scope[i]))
			}

			if scope[i] == Wildcard && i < len(scope)-1 && scope[i+1] == Wildcard && i < len(scope)-2 {
				return errSuperNotLast
			}
		}
	}

	return nil
}

func compareRuleToScope(
	rule *string,
	scope *string,
	vars map[string]string,
) (bool, error) {
	// Skip the allow and deny prefix for actors
	ruleLeft, _, _ := endOfBlock(rule, 0, "rule")

	ruleLeft += 1 // don't forget to skip the slash
	scopeLeft := 0

	for ruleLeft < len(*rule) || scopeLeft < len(*scope) {
		// In case one is longer then the other
		if (ruleLeft < len(*rule)) != (scopeLeft < len(*scope)) {
			return false, nil
		}

		scopeSlider, _, err := endOfBlock(scope, scopeLeft, "scope")
		if err != nil {
			return false, err
		}

		ruleSlider, ruleArray, err := endOfBlock(rule, ruleLeft, "rule")
		if err != nil {
			return false, err
		}

		// Super wildcards are checked here as it skips the who rest of the checks.
		if ruleSlider-ruleLeft == 2 && (*rule)[ruleLeft] == Wildcard && (*rule)[ruleLeft+1] == Wildcard {
			if len(*rule) > ruleSlider {
				return false, errSuperNotLast
			}

			return true, nil
		} else {
			match, err := compareBlock(rule, ruleLeft, ruleSlider, ruleArray, scope, scopeLeft, scopeSlider, vars)
			if err != nil {
				return false, err
			}

			if !match {
				return false, nil
			}
		}

		scopeLeft = scopeSlider + 1
		ruleLeft = ruleSlider + 1
	}

	return true, nil
}

func compareBlock(
	rule *string, ruleLeft, ruleSlider int, ruleArray bool,
	scope *string, scopeLeft, scopeSlider int,
	vars map[string]string,
) (bool, error) {
	if (*rule)[ruleLeft] == VariablePrefix {
		key := (*rule)[ruleLeft+1 : ruleSlider]
		varValue, found := vars[key]

		if !found {
			return false, fmt.Errorf(fmtAllowedVarNotFound, key)
		}

		return varValue == (*scope)[scopeLeft:scopeSlider], nil
	}

	if ruleSlider-ruleLeft == 1 && (*rule)[ruleLeft] == Wildcard {
		return true, nil
	}

	if ruleArray {
		for ruleLeft < ruleSlider {
			arrayRight := endOfArrayElement(rule, ruleLeft)

			if (*rule)[ruleLeft] == VariablePrefix {
				key := (*rule)[ruleLeft+1 : arrayRight]
				return false, fmt.Errorf(fmtAllowedVarInArray, key)
			}

			if (*rule)[ruleLeft] == Wildcard {
				if arrayRight-ruleLeft > 1 && (*rule)[ruleLeft+1] == Wildcard {
					return false, errSuperInArray
				}

				return false, errWildcardInArray
			}

			if (*rule)[ruleLeft:arrayRight] == (*scope)[scopeLeft:scopeSlider] {
				return true, nil
			}

			ruleLeft = arrayRight + 1
		}

		return false, nil
	}

	return (*rule)[ruleLeft:ruleSlider] == (*scope)[scopeLeft:scopeSlider], nil
}

func endOfBlock(value *string, start int, category string) (int, bool, error) {
	isArray := false

	for i := start; i < len(*value); i++ {
		if (*value)[i] == ArraySeperator {
			isArray = true
		} else if (*value)[i] == BlockSeperator {
			return i, isArray, nil
		} else if !isValidCharacter((*value)[i]) {
			invalidChar := string((*value)[i])
			return 0, false, fmt.Errorf(fmtAllowedInvalidChar, category, invalidChar)
		}
	}

	return len(*value), isArray, nil
}

func endOfArrayElement(value *string, start int) int {
	for i := start + 1; i < len(*value); i++ {
		if (*value)[i] == BlockSeperator ||
			(*value)[i] == ArraySeperator {
			return i
		}
	}

	return len(*value)
}

func isValidCharacter(char byte) bool {
	if char >= 'a' && char <= 'z' {
		return true
	}

	if char >= 'A' && char <= 'Z' {
		return true
	}

	if char >= '0' && char <= '9' {
		return true
	}

	return char == '_' || char == '-' || char == VariablePrefix || char == Wildcard
}
