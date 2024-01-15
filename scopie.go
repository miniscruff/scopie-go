package scopie

import "fmt"

const (
	BlockSeperator = byte('/')
	ScopeSeperator = byte(',')
	ArraySeperator = byte('|')
	VariablePrefix = byte('@')
	Wildcard       = byte('*')
)

// IsAllowed returns whether or not the required role scopes are fulfilled by our actor scopes.
func IsAllowed(vars map[string]string, requiredScopes, actorScopes string) (bool, error) {
	if requiredScopes == "" {
		return false, fmt.Errorf("scopie-106 in scopes@0: scope was empty")
	}

	if actorScopes == "" {
		return false, fmt.Errorf("scopie-106 in actor@0: scope was empty")
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
			return 0, 0, false, fmt.Errorf("scopie-105 in actor@%d: super wildcard not in the last block", aIndex)
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
			err := fmt.Errorf("scopie-100 in scopes@%d: invalid character '%s'", bSlider, invalidChar)
			return 0, 0, false, err
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
			if (*aValue)[aLeft] == '@' {
				return 0, 0, false, fmt.Errorf("scopie-101 in actor@%d: variable '%s' found in array block", aLeft, (*aValue)[aLeft+1:aSlider])
			}

			if (*aValue)[aLeft] == '*' {
				if aLeft < len(*aValue)-1 && (*aValue)[aLeft+1] == '*' {
					return 0, 0, false, fmt.Errorf("scopie-103 in actor@%d: super wildcard found in array block", aLeft)
				}

				return 0, 0, false, fmt.Errorf("scopie-102 in actor@%d: wildcard found in array block", aLeft)
			}

			match, _ := compareChunk(aValue, aLeft, aSlider, bValue, bIndex, bSlider, nil)
			if match {
				return jumpBlockOrScopeSeperator(aValue, aSlider), bSlider + 1, true, nil
			}

			// go to the next array value
			aLeft = aSlider + 1
			aSlider += 1
		} else if !isValidCharacter((*aValue)[aSlider]) {
			return 0, 0, false, fmt.Errorf("scopie-100 in actor@%d: invalid character '%s'", aSlider, string((*aValue)[aSlider]))
		}
	}

	if wasArray {
		if (*aValue)[aLeft] == '@' {
			return 0, 0, false, fmt.Errorf("scopie-101 in actor@%d: variable '%s' found in array block", aLeft, (*aValue)[aLeft+1:aSlider])
		}

		if (*aValue)[aLeft] == '*' {
			if aLeft < len(*aValue)-1 && (*aValue)[aLeft+1] == '*' {
				return 0, 0, false, fmt.Errorf("scopie-103 in actor@%d: super wildcard found in array block", aLeft)
			}

			return 0, 0, false, fmt.Errorf("scopie-102 in actor@%d: wildcard found in array block", aLeft)
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

func compareChunk(aValue *string, aLeft, aSlider int, bValue *string, bLeft, bSlider int, vars map[string]string) (bool, error) {
	if (*aValue)[aLeft] == '@' {
		key := (*aValue)[aLeft+1 : aSlider]
		varValue, found := vars[key]

		if !found {
			return false, fmt.Errorf("scopie-104 in actor@%d: variable '%s' not found", aLeft, key)
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

	return char == '_' || char == '-' || char == '@' || char == '*'
}
