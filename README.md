# Scopie

not prod ready

## Basic Example
```go
import (
    "log/slog"
    scopie "github.com/miniscruff/scopie-go"
)

func main() {
    result, err := scopie.Process(
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
    if result == scopie.ResultDeny {
        slog.Error("unauthorized")
        return
    }

    // User is allowed to create a blog post
}
```
