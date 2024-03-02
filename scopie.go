package scopie

import "fmt"

const (
	BlockSeperator = byte('/')
	ScopeSeperator = byte(',')
	ArraySeperator = byte('|')
	VariablePrefix = byte('@')
	Wildcard       = byte('*')
)

const (
	fmtAllowedErrInvalidChar     = "scopie-100 in %s@%d: invalid character '%s'"
	fmtAllowedErrVarInArray      = "scopie-101 in actor@%d: variable '%s' found in array block"
	fmtAllowedErrWildcardInArray = "scopie-102 in actor@%d: wildcard found in array block"
	fmtAllowedErrSuperInArray    = "scopie-103 in actor@%d: super wildcard found in array block"
	fmtAllowedErrVarNotFound     = "scopie-104 in actor@%d: variable '%s' not found"
	fmtAllowedErrSuperNotLast    = "scopie-105 in actor@%d: super wildcard not in the last block"
	fmtAllowedErrEmpty           = "scopie-106 in %s@0: scope was empty"

	fmtValidateErrInvalidChar     = "scopie-100@%d: invalid character '%s'"
	fmtValidateErrVarInArray      = "scopie-101@%d: variable '%s' found in array block"
	fmtValidateErrWildcardInArray = "scopie-102@%d: wildcard found in array block"
	fmtValidateErrSuperInArray    = "scopie-103@%d: super wildcard found in array block"
	fmtValidateErrSuperNotLast    = "scopie-105@%d: super wildcard not in the last block"
	fmtValidateErrEmpty           = "scopie-106@%d: scope was empty"
)

// IsAllowedFunc is a type wrapper for IsAllowed that can be used as
// a dependency.
type IsAllowedFunc func(map[string]string, string, string) (bool, error)

// ValidateScopeFunc is a type wrapper for ValidateScope that can be
// used as a dependency.
type ValidateScopeFunc func(string) error

// IsAllowed returns whether or not the required role scopes are fulfilled by our actor scopes.
func IsAllowed(vars map[string]string, requiredScopes, actorScopes string) (bool, error) {
	if requiredScopes == "" {
		return false, fmt.Errorf(fmtAllowedErrEmpty, "scopes")
	}

	if actorScopes == "" {
		return false, fmt.Errorf(fmtAllowedErrEmpty, "actor")
	}

	actorIndex := 0
	actorLeft := 0
	hasBeenAllowed := false

	for actorLeft < len(actorScopes) {
		isAllowBlock := actorScopes[actorLeft] == 'a'
		if isAllowBlock && hasBeenAllowed {
			actorLeft = jumpAfterSeperator(&actorScopes, actorLeft, ScopeSeperator)
			continue
		}

		actorLeft = jumpAfterSeperator(&actorScopes, actorLeft, BlockSeperator)
		actorIndex = actorLeft
		ruleLeft := 0

		for ruleLeft < len(requiredScopes) {
			actorNext, ruleNext, matched, err := compareFrom(&actorScopes, actorLeft, &requiredScopes, ruleLeft, vars)
			if err != nil {
				return false, err
			}

			if matched {
				actorLeft = actorNext
				ruleLeft = ruleNext

				endOfActor := actorLeft >= len(actorScopes) || actorScopes[actorLeft-1] == ScopeSeperator
				endOfRequired := ruleLeft >= len(requiredScopes) || requiredScopes[ruleLeft-1] == ScopeSeperator

				// if we are at the end of the actor and of the required scope
				if endOfActor && endOfRequired {
					if isAllowBlock {
						hasBeenAllowed = true
						actorLeft = jumpAfterSeperator(&actorScopes, actorLeft, ScopeSeperator)
					} else {
						return false, nil
					}

					break
				} else if endOfActor != endOfRequired {
					break
				}
			} else {
				ruleLeft = jumpAfterSeperator(&requiredScopes, ruleLeft, ScopeSeperator)
				actorLeft = actorIndex
			}
		}

		actorLeft = jumpAfterSeperator(&actorScopes, actorLeft, ScopeSeperator)
	}

	return hasBeenAllowed, nil
}

func ValidateScope(scope string) error {
	if scope == "" {
		return fmt.Errorf(fmtValidateErrEmpty, 0)
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
				return fmt.Errorf(fmtValidateErrSuperInArray, i)
			}

			if scope[i] == Wildcard {
				return fmt.Errorf(fmtValidateErrWildcardInArray, i)
			}

			if scope[i] == VariablePrefix {
				end := jumpEndOfArrayElement(&scope, i)
				return fmt.Errorf(fmtValidateErrVarInArray, i, scope[i+1:end])
			}
		}

		if !isValidCharacter(scope[i]) {
			return fmt.Errorf(fmtValidateErrInvalidChar, i, string(scope[i]))
		}

		if scope[i] == Wildcard && i < len(scope)-1 && scope[i+1] == Wildcard &&
			i < len(scope)-2 && scope[i+2] != ScopeSeperator {
			return fmt.Errorf(fmtValidateErrSuperNotLast, i)
		}
	}

	return nil
}

func compareFrom(
	aValue *string,
	aIndex int,
	bValue *string,
	bIndex int,
	vars map[string]string,
) (int, int, bool, error) {
	// Super wildcard is just two wildcards
	if (*aValue)[aIndex] == Wildcard && aIndex < len(*aValue)-1 && (*aValue)[aIndex+1] == Wildcard {
		if aIndex+2 < len(*aValue) && (*aValue)[aIndex+2] != ScopeSeperator {
			return 0, 0, false, fmt.Errorf(fmtAllowedErrSuperNotLast, aIndex)
		}

		newAIndex := jumpAfterSeperator(aValue, aIndex, ScopeSeperator)
		newBIndex := jumpAfterSeperator(bValue, bIndex, ScopeSeperator)

		return newAIndex, newBIndex, true, nil
	}

	if (*aValue)[aIndex] == Wildcard {
		newAIndex := jumpAfterSeperator(aValue, aIndex, BlockSeperator)
		newBIndex := jumpAfterSeperator(bValue, bIndex, BlockSeperator)

		return newAIndex, newBIndex, true, nil
	}

	bSlider := bIndex
	for ; bSlider < len(*bValue); bSlider++ {
		if (*bValue)[bSlider] == BlockSeperator || (*bValue)[bSlider] == ScopeSeperator {
			break
		} else if !isValidCharacter((*bValue)[bSlider]) {
			invalidChar := string((*bValue)[bSlider])
			return 0, 0, false, fmt.Errorf(fmtAllowedErrInvalidChar, "scopes", bSlider, invalidChar)
		}
	}

	aLeft := aIndex
	aSlider := aIndex
	wasArray := false

	for ; aSlider < len(*aValue); aSlider++ {
		if (*aValue)[aSlider] == BlockSeperator || (*aValue)[aSlider] == ScopeSeperator {
			match, err := compareChunk(aValue, aLeft, aSlider, bValue, bIndex, bSlider, vars)
			if err != nil {
				return 0, 0, false, err
			}

			if match {
				return aSlider + 1, bSlider + 1, true, nil
			}

			return aIndex, bIndex, false, nil
		} else if (*aValue)[aSlider] == ArraySeperator {
			wasArray = true

			if (*aValue)[aLeft] == VariablePrefix {
				return 0, 0, false, fmt.Errorf(fmtAllowedErrVarInArray, aLeft, (*aValue)[aLeft+1:aSlider])
			}

			if (*aValue)[aLeft] == Wildcard {
				if aLeft < len(*aValue)-1 && (*aValue)[aLeft+1] == Wildcard {
					return 0, 0, false, fmt.Errorf(fmtAllowedErrSuperInArray, aLeft)
				}

				return 0, 0, false, fmt.Errorf(fmtAllowedErrWildcardInArray, aLeft)
			}

			match, _ := compareChunk(aValue, aLeft, aSlider, bValue, bIndex, bSlider, nil)
			if match {
				return jumpBlockOrScopeSeperator(aValue, aSlider), bSlider + 1, true, nil
			}

			// go to the next array value
			aLeft = aSlider + 1
			aSlider += 1
		} else if !isValidCharacter((*aValue)[aSlider]) {
			return 0, 0, false, fmt.Errorf(fmtAllowedErrInvalidChar, "actor", aSlider, string((*aValue)[aSlider]))
		}
	}

	if wasArray {
		if (*aValue)[aLeft] == VariablePrefix {
			return 0, 0, false, fmt.Errorf(fmtAllowedErrVarInArray, aLeft, (*aValue)[aLeft+1:aSlider])
		}

		if (*aValue)[aLeft] == Wildcard {
			if aLeft < len(*aValue)-1 && (*aValue)[aLeft+1] == Wildcard {
				return 0, 0, false, fmt.Errorf(fmtAllowedErrSuperInArray, aLeft)
			}

			return 0, 0, false, fmt.Errorf(fmtAllowedErrWildcardInArray, aLeft)
		}
	}

	match, err := compareChunk(aValue, aLeft, aSlider, bValue, bIndex, bSlider, vars)
	if err != nil {
		return 0, 0, false, err
	}

	if match {
		return aSlider + 1, bSlider + 1, true, nil
	}

	return aIndex, bIndex, false, nil
}

func compareChunk(
	aValue *string, aLeft, aSlider int,
	bValue *string, bLeft, bSlider int,
	vars map[string]string,
) (bool, error) {
	if (*aValue)[aLeft] == VariablePrefix {
		key := (*aValue)[aLeft+1 : aSlider]
		varValue, found := vars[key]

		if !found {
			return false, fmt.Errorf(fmtAllowedErrVarNotFound, aLeft, key)
		}

		return varValue == (*bValue)[bLeft:bSlider], nil
	}

	if aSlider-aLeft != bSlider-bLeft {
		return false, nil
	}

	return (*aValue)[aLeft:aSlider] == (*bValue)[bLeft:bSlider], nil
}

func jumpAfterSeperator(value *string, start int, sep byte) int {
	for i := start + 1; i < len(*value); i++ {
		if (*value)[i] == sep {
			return i + 1
		}
	}

	return len(*value)
}

func jumpBlockOrScopeSeperator(value *string, start int) int {
	for i := start + 1; i < len(*value); i++ {
		if (*value)[i] == BlockSeperator || (*value)[i] == ScopeSeperator {
			return i + 1
		}
	}

	return len(*value)
}

func jumpEndOfArrayElement(value *string, start int) int {
	for i := start + 1; i < len(*value); i++ {
		if (*value)[i] == BlockSeperator ||
			(*value)[i] == ScopeSeperator ||
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
