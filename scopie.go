package scopie

import (
	"log/slog"
	"slices"
	"strings"
)

const (
	BlockSeperator = "/"
	ScopeSeperator = ","
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
	logger.Info("processing scopes")

	for _, actorScope := range strings.Split(actorScopes, ScopeSeperator) {
		actorSplit := strings.Split(actorScope, BlockSeperator)
		actorScopes := actorSplit[1:]
		rule := actorSplit[0]

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
