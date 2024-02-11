# Scopie

_not production ready_

Go implementation of [scopie](https://github.com/miniscruff/scopie).

## Basic Example
```go
import (
    "log/slog"
    scopie "github.com/miniscruff/scopie-go"
)

func main() {
    allowed, err := scopie.IsAllowed(
        // optional variable values if we used any @vars
        map[string]string{},
        // an example user scope
        "allow/blog/post/create",
        // what our request requires
        "blog/post/create",
    )
    // If there was an error parsing a scope or rule.
    if err != nil {
        slog.Error("processing scopes", "error", err.Error())
        return
    }

    // Check the result to see if we can do this action.
    if !allowed {
        slog.Error("unauthorized")
        return
    }

    // User is allowed to create a blog post
}
```
