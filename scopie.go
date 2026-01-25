package scopie

import (
	"errors"
	"fmt"
	"strings"
)

const (
	BlockSeparator = byte('/')
	ArraySeparator = byte('|')
	VariablePrefix = byte('@')
	Wildcard       = byte('*')

	AllowGrant     = "allow"
	DenyGrant      = "deny"
	GrantSeparator = byte(':')
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
	errActionsEmpty    = errors.New("scopie-106 in action: actions was empty")
	errActionEmpty     = errors.New("scopie-106 in action: action was empty")
	errPermissionEmpty = errors.New("scopie-106 in permission: permission was empty")

	errPermissionDoesNotStartWithGrant = errors.New(
		"scopie-107: permission does not start with a grant",
	)

	// validation specific
	errValidateActionEmpty      = errors.New("scopie-106: action was empty")
	errValidateActionsEmpty     = errors.New("scopie-106: action array was empty")
	errValidatePermissionEmpty  = errors.New("scopie-106: permission was empty")
	errValidatePermissionsEmpty = errors.New("scopie-106: permission array was empty")
)

// IsAllowedFunc is a type wrapper for [IsAllowed] to be used a dependency.
type IsAllowedFunc func(map[string]string, string, string) (bool, error)

// ValidateActionsFunc is a type wrapper for [ValidateActions] to be used as a dependency.
type ValidateActionsFunc func(actions []string) error

// ValidateActionsFunc is a type wrapper for [ValidatePermissions] to be used as a dependency.
type ValidatePermissionsFunc func(permissions []string) error

// IsAllowed returns whether or not the user with the given permissions are allowed to complete
// the action. See [Is Allowed Spec] for additional details.
//
//	isAllowed, err := IsAllowed(
//		[]string{"accounts/thor/edit"},
//		"allow:accounts/@username/*",
//		map[string]string{"username": "thor"},
//	)
//	if err != nil {
//		return fmt.Errorf("invalid action or permission: %w", err)
//	}
//	if !isAllowed {
//		return fmt.Errorf("unauthorized")
//	}
//
// [Is Allowed Spec]: https://scopie.dev/specification/functions/#is-allowed
func IsAllowed(actions, permissions []string, vars map[string]string) (bool, error) {
	if len(actions) == 0 {
		return false, errActionsEmpty
	}

	if len(permissions) == 0 {
		return false, nil
	}

	hasBeenAllowed := false

	for _, permission := range permissions {
		if len(permission) == 0 {
			return false, errPermissionEmpty
		}

		isAllowBlock := strings.HasPrefix(permission, AllowGrant)
		if !isAllowBlock && !strings.HasPrefix(permission, DenyGrant) {
			return false, errPermissionDoesNotStartWithGrant
		}

		if isAllowBlock && hasBeenAllowed {
			continue
		}

		for _, action := range actions {
			if len(action) == 0 {
				return false, errActionEmpty
			}

			match, err := comparePermissionToAction(&permission, &action, vars)
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

// ValidateActions checks whether or not the given actions are valid given the
// requirements outlined in the specification.
// [Validate Scopes Spec] is the function specification.
//
//	err := ValidateActions("accounts/create")
//	if err != nil {
//		return fmt.Errorf("action is invalid: %w", err)
//	}
//
// [Validate Actions Spec]: https://scopie.dev/specification/functions/#validate-actions
func ValidateActions(actions ...string) error {
	if len(actions) == 0 {
		return errValidateActionsEmpty
	}

	for _, action := range actions {
		if action == "" {
			return errValidateActionEmpty
		}

		for i := range action {
			if action[i] == BlockSeparator {
				continue
			}

			if !isValidLiteral(action[i]) {
				return fmt.Errorf(fmtValidateInvalidChar, string(action[i]))
			}
		}
	}

	return nil
}

// ValidatePermissions checks whether or not the given permissions are valid given the
// requirements outlined in the specification.
// [Validate Permissions Spec] is the function specification.
//
//	err := ValidatePermissionsFunc("allow:accounts/read")
//	if err != nil {
//		return fmt.Errorf("action is invalid: %w", err)
//	}
//
// [Validate Permissions Spec]: https://scopie.dev/specification/functions/#validate-permissions
func ValidatePermissions(permissions ...string) error {
	if len(permissions) == 0 {
		return errValidatePermissionsEmpty
	}

	for _, permission := range permissions {
		if permission == "" {
			return errValidatePermissionEmpty
		}

		inArray := false

		i, err := skipGrant(&permission, 0)
		if err != nil {
			return errPermissionDoesNotStartWithGrant
		}

		for ; i < len(permission); i++ {
			if permission[i] == BlockSeparator {
				inArray = false
				continue
			}

			if permission[i] == ArraySeparator {
				inArray = true
				continue
			}

			if inArray {
				if permission[i] == Wildcard && i < len(permission)-1 &&
					permission[i+1] == Wildcard {
					return errSuperInArray
				}

				if permission[i] == Wildcard {
					return errWildcardInArray
				}

				if permission[i] == VariablePrefix {
					end := endOfArrayElement(&permission, i)
					return fmt.Errorf(fmtValidateVarInArray, permission[i+1:end])
				}
			}

			if !isValidCharacter(permission[i]) {
				return fmt.Errorf(fmtValidateInvalidChar, string(permission[i]))
			}

			if permission[i] == Wildcard && i < len(permission)-1 && permission[i+1] == Wildcard &&
				i < len(permission)-2 {
				return errSuperNotLast
			}
		}
	}

	return nil
}

func comparePermissionToAction(
	permission *string,
	action *string,
	vars map[string]string,
) (bool, error) {
	// skip grant error is pre-checked
	permissionLeft, _ := skipGrant(permission, 0)
	actionLeft := 0

	for permissionLeft < len(*permission) || actionLeft < len(*action) {
		// In case one is longer then the other
		if (permissionLeft < len(*permission)) != (actionLeft < len(*action)) {
			return false, nil
		}

		actionSlider, _, err := endOfBlock(action, actionLeft, "action")
		if err != nil {
			return false, err
		}

		permissionSlider, permissionArray, err := endOfBlock(
			permission,
			permissionLeft,
			"permission",
		)
		if err != nil {
			return false, err
		}

		// Super wildcards are checked here as it skips the who rest of the checks.
		if permissionSlider-permissionLeft == 2 && (*permission)[permissionLeft] == Wildcard &&
			(*permission)[permissionLeft+1] == Wildcard {
			if len(*permission) > permissionSlider {
				return false, errSuperNotLast
			}

			return true, nil
		} else {
			match, err := compareBlock(
				permission,
				permissionLeft,
				permissionSlider,
				permissionArray,
				action,
				actionLeft,
				actionSlider,
				vars,
			)
			if err != nil {
				return false, err
			}

			if !match {
				return false, nil
			}
		}

		actionLeft = actionSlider + 1
		permissionLeft = permissionSlider + 1
	}

	return true, nil
}

func compareBlock(
	permission *string, permissionLeft, permissionSlider int, permissionArray bool,
	action *string, actionLeft, actionSlider int,
	vars map[string]string,
) (bool, error) {
	if (*permission)[permissionLeft] == VariablePrefix {
		key := (*permission)[permissionLeft+1 : permissionSlider]
		varValue, found := vars[key]

		if !found {
			return false, fmt.Errorf(fmtAllowedVarNotFound, key)
		}

		return varValue == (*action)[actionLeft:actionSlider], nil
	}

	if permissionSlider-permissionLeft == 1 && (*permission)[permissionLeft] == Wildcard {
		return true, nil
	}

	if permissionArray {
		for permissionLeft < permissionSlider {
			arrayRight := endOfArrayElement(permission, permissionLeft)

			if (*permission)[permissionLeft] == VariablePrefix {
				key := (*permission)[permissionLeft+1 : arrayRight]
				return false, fmt.Errorf(fmtAllowedVarInArray, key)
			}

			if (*permission)[permissionLeft] == Wildcard {
				if arrayRight-permissionLeft > 1 && (*permission)[permissionLeft+1] == Wildcard {
					return false, errSuperInArray
				}

				return false, errWildcardInArray
			}

			if (*permission)[permissionLeft:arrayRight] == (*action)[actionLeft:actionSlider] {
				return true, nil
			}

			permissionLeft = arrayRight + 1
		}

		return false, nil
	}

	return (*permission)[permissionLeft:permissionSlider] == (*action)[actionLeft:actionSlider], nil
}

func skipGrant(value *string, start int) (int, error) {
	subStr := (*value)[start:]

	if strings.HasPrefix(subStr, DenyGrant) {
		return 5, nil
	}

	if strings.HasPrefix(subStr, AllowGrant) {
		return 6, nil
	}

	return 0, errPermissionDoesNotStartWithGrant
}

func endOfBlock(value *string, start int, category string) (int, bool, error) {
	isArray := false

	for i := start; i < len(*value); i++ {
		if (*value)[i] == ArraySeparator {
			isArray = true
		} else if (*value)[i] == BlockSeparator {
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
		if (*value)[i] == BlockSeparator || (*value)[i] == ArraySeparator {
			return i
		}
	}

	return len(*value)
}

func isValidLiteral(char byte) bool {
	if char >= 'a' && char <= 'z' {
		return true
	}

	if char >= 'A' && char <= 'Z' {
		return true
	}

	if char >= '0' && char <= '9' {
		return true
	}

	return char == '_' || char == '-'
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
