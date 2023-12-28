package scopie

const (
	BlockSeperator = byte('/')
	ScopeSeperator = byte(',')
	ArraySeperator = byte('|')
	VariablePrefix = byte('@')
	Wildcard       = byte('*')
)

// IsAllowed returns whether or not the required role scopes are fulfilled by our actor scopes.
func IsAllowed(vars map[string]string, requiredRules, actorScopes string) (bool, error) {
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

		for ruleLeft < len(requiredRules) {
			actorNext, ruleNext, matched := compareFrom(&actorScopes, actorLeft, &requiredRules, ruleLeft, vars)
			if matched {
				actorLeft = actorNext
				ruleLeft = ruleNext

				// if we are at the end of the actor...
				if actorLeft >= len(actorScopes) || actorScopes[actorLeft-1] == ScopeSeperator {
					if isAllowBlock {
						hasBeenAllowed = true
						actorLeft = jumpAfterSeperator(&actorScopes, actorLeft, ScopeSeperator)
					} else {
						return false, nil
					}
					break
				}
			} else {
				ruleLeft = jumpAfterSeperator(&requiredRules, ruleLeft, ScopeSeperator)
				actorLeft = actorIndex
			}
		}

		actorLeft = jumpAfterSeperator(&actorScopes, actorLeft, ScopeSeperator)
	}

	return hasBeenAllowed, nil
}

func compareFrom(aValue *string, aIndex int, bValue *string, bIndex int, vars map[string]string) (int, int, bool) {
	// Super wildcard is just two wildcards
	if (*aValue)[aIndex] == Wildcard && (*aValue)[aIndex+1] == Wildcard {
		return jumpAfterSeperator(aValue, aIndex, ScopeSeperator), jumpAfterSeperator(bValue, bIndex, ScopeSeperator), true
	}

	if (*aValue)[aIndex] == Wildcard {
		return jumpAfterSeperator(aValue, aIndex, BlockSeperator), jumpAfterSeperator(bValue, bIndex, BlockSeperator), true
	}

	bSlider := bIndex
	for ; bSlider < len(*bValue); bSlider++ {
		if (*bValue)[bSlider] == BlockSeperator || (*bValue)[bSlider] == ScopeSeperator {
			break
		}
	}

	aLeft := aIndex
	aSlider := aIndex
	for ; aSlider < len(*aValue); aSlider++ {
		if (*aValue)[aSlider] == BlockSeperator || (*aValue)[aSlider] == ScopeSeperator {
			if compareChunk(aValue, aLeft, aSlider, bValue, bIndex, bSlider, vars) {
				return aSlider + 1, bSlider + 1, true
			}

			return aIndex, bIndex, false
		} else if (*aValue)[aSlider] == ArraySeperator {
			if compareChunk(aValue, aLeft, aSlider, bValue, bIndex, bSlider, nil) {
				return jumpBlockOrScopeSeperator(aValue, aSlider), bSlider + 1, true
			}

			// go to the next array value
			aLeft = aSlider + 1
			aSlider += 2
		}
	}

	if compareChunk(aValue, aLeft, aSlider, bValue, bIndex, bSlider, vars) {
		return aSlider + 1, bSlider + 1, true
	}

	return aIndex, bIndex, false
}

func compareChunk(aValue *string, aLeft, aSlider int, bValue *string, bLeft, bSlider int, vars map[string]string) bool {
	if (*aValue)[aLeft] == '@' {
		return vars[(*aValue)[aLeft+1:aSlider]] == (*bValue)[bLeft:bSlider]
	}

	if aSlider-aLeft != bSlider-bLeft {
		return false
	}

	return (*aValue)[aLeft:aSlider] == (*bValue)[bLeft:bSlider]
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
