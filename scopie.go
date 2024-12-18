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
	errSuperNotLast      = errors.New("scopie-105: super wildcard not in the last block")
	errSuperInArray      = errors.New("scopie-103: super wildcard found in array block")
	errWildcardInArray   = errors.New("scopie-102: wildcard found in array block")
	errActionScopesEmpty = errors.New("scopie-106 in action: scopes was empty")
	errActionScopeEmpty  = errors.New("scopie-106 in action: scope was empty")
	errActorRuleEmpty    = errors.New("scopie-106 in actor: rule was empty")

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
func IsAllowed(actionScopes, actorRules []string, vars map[string]string) (bool, error) {
	if len(actionScopes) == 0 {
		return false, errActionScopesEmpty
	}

	if len(actorRules) == 0 {
		return false, nil
	}

	hasBeenAllowed := false

	for _, actorRule := range actorRules {
		if len(actorRule) == 0 {
			return false, errActorRuleEmpty
		}

		actorRule := actorRule

		isAllowBlock := strings.HasPrefix(actorRule, AllowPermission)
		if isAllowBlock && hasBeenAllowed {
			continue
		}

		for _, actionScope := range actionScopes {
			if len(actionScope) == 0 {
				return false, errActionScopeEmpty
			}

			actionScope := actionScope

			match, err := compareActorToAction(&actorRule, &actionScope, vars)
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

func compareActorToAction(
	actor *string,
	action *string,
	vars map[string]string,
) (bool, error) {
	// Skip the allow and deny prefix for actors
	actorLeft, _, _ := endOfBlock(actor, 0, "actor")

	actorLeft += 1 // don't forget to skip the slash
	actionLeft := 0

	for actorLeft < len(*actor) || actionLeft < len(*action) {
		// In case one is longer then the other
		if (actorLeft < len(*actor)) != (actionLeft < len(*action)) {
			return false, nil
		}

		actionSlider, _, err := endOfBlock(action, actionLeft, "action")
		if err != nil {
			return false, err
		}

		actorSlider, actorArray, err := endOfBlock(actor, actorLeft, "actor")
		if err != nil {
			return false, err
		}

		// Super wildcards are checked here as it skips the who rest of the checks.
		if actorSlider-actorLeft == 2 && (*actor)[actorLeft] == Wildcard && (*actor)[actorLeft+1] == Wildcard {
			if len(*actor) > actorSlider {
				return false, errSuperNotLast
			}

			return true, nil
		} else {
			match, err := compareBlock(actor, actorLeft, actorSlider, actorArray, action, actionLeft, actionSlider, vars)
			if err != nil {
				return false, err
			}

			if !match {
				return false, nil
			}
		}

		actionLeft = actionSlider + 1
		actorLeft = actorSlider + 1
	}

	return true, nil
}

func compareBlock(
	actor *string, actorLeft, actorSlider int, actorArray bool,
	action *string, actionLeft, actionSlider int,
	vars map[string]string,
) (bool, error) {
	if (*actor)[actorLeft] == VariablePrefix {
		key := (*actor)[actorLeft+1 : actorSlider]
		varValue, found := vars[key]

		if !found {
			return false, fmt.Errorf(fmtAllowedVarNotFound, key)
		}

		return varValue == (*action)[actionLeft:actionSlider], nil
	}

	if actorSlider-actorLeft == 1 && (*actor)[actorLeft] == Wildcard {
		return true, nil
	}

	if actorArray {
		for actorLeft < actorSlider {
			arrayRight := endOfArrayElement(actor, actorLeft)

			if (*actor)[actorLeft] == VariablePrefix {
				key := (*actor)[actorLeft+1 : arrayRight]
				return false, fmt.Errorf(fmtAllowedVarInArray, key)
			}

			if (*actor)[actorLeft] == Wildcard {
				if arrayRight-actorLeft > 1 && (*actor)[actorLeft+1] == Wildcard {
					return false, errSuperInArray
				}

				return false, errWildcardInArray
			}

			if (*actor)[actorLeft:arrayRight] == (*action)[actionLeft:actionSlider] {
				return true, nil
			}

			actorLeft = arrayRight + 1
		}

		return false, nil
	}

	return (*actor)[actorLeft:actorSlider] == (*action)[actionLeft:actionSlider], nil
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
