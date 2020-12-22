# Go Nested Set

[![build](https://github.com/griffinqiu/go-nested-set/workflows/build/badge.svg)](https://github.com/griffinqiu/go-nested-set/actions?query=workflow%3Abuild)

Go Nested Set is an implementation of the [Nested set model](https://en.wikipedia.org/wiki/Nested_set_model) for [Gorm](https://gorm.io/index.html).

This project is the Go version of [awesome_nested_set](https://github.com/collectiveidea/awesome_nested_set), which uses the same data structure design, so it uses the same data together with [awesome_nested_set](https://github.com/collectiveidea/awesome_nested_set).

> Actually the original design is for this, the content managed by [awesome_nested_set](https://github.com/collectiveidea/awesome_nested_set) in our Rails application, the front-end Go API also needs to be maintained.

This is a Go version of the [awesome_nested_set](https://github.com/collectiveidea/awesome_nested_set), and it built for compatible with [awesome_nested_set](https://github.com/collectiveidea/awesome_nested_set), so you can use

## Installation

```
go get github.com/griffinqiu/go-nested-set
```

## Usage

### Define the model

You must use `nestedset` Stuct tag to define your Gorm model like this:

Support struct tags:

- `id` - int64 - Primary key of the node
- `parent_id` - int64 - ParentID column, 0 is root
- `lft` - int
- `rgt` - int
- `depth` - int - Depth of the node
- `children_count` - Number of children

Example:

```go
import "github.com/griffinqiu/go-nested-set"

// Category
type Category struct {
	ID            int64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT" nestedset:"id"`
	Title         string
	ParentID      int64 `nestedset:"parent_id"`
	Rgt           int   `nestedset:"rgt"`
	Lft           int   `nestedset:"lft"`
	Depth         int   `nestedset:"depth"`
	ChildrenCount int   `nestedset:"children_count"`
}
```

### Move Node

```go
import nestedset "github.com/griffinqiu/go-nested-set"

// nestedset.MoveDirectionLeft
// nestedset.MoveDirectionRight
// nestedset.MoveDirectionInner

nestedset.MoveTo(gormDB, node, to, nestedset.MoveDirectionLeft)
```
