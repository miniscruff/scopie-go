# Scopie

[![Go Reference](https://pkg.go.dev/badge/github.com/miniscruff/scopie-go.svg)](https://pkg.go.dev/github.com/miniscruff/scopie-go)

Go implementation of [scopie](https://github.com/miniscruff/scopie).

## Example

```go
import (
    "errors"
    "github.com/miniscruff/scopie-go"
)

type User struct {
    Scopes []string
}

type BlogPost struct {
    Author  User
    Content string
}

var userStore map[string]User = map[string]User{
    "elsa": User{
        Scopes: []string{"allow/blog/create|update"},
    },
    "belle": User{
        Scopes: []string{"allow/blog/create"},
    },
}
var blogStore map[string]BlogPost = map[string]BlogPost{}

func createBlog(username, blogSlug, blogContent string) error {
    user := users[username]
    allowed, err := scopie.IsAllowed([]string{"blog/create"}, user.scopes, nil)
    if err != nil {
        return err
    }

    if !allowed {
        return errors.New("not allowed to create a blog post")
    }

    blogStore[blogSlug] = BlogPost{
        Author: user,
        Content: blogContent,
    }
    return nil
}

func updateBlog(username, blogSlug, blogContent string) error {
    user := users[username]
    allowed, err := scopie.IsAllowed([]string{"blog/update"}, user.scopes, nil) {
    if err != nil {
        return err
    }

    if !allowed {
        return errors.New("not allowed to update this blog post")
    }

    blogPosts[blogSlug] = BlogPost{
        author: user,
        content: blogContent,
    }
    return nil
}
```
