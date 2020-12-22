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
- `parent_id` - int64 - ParentID column, null is root
- `lft` - int
- `rgt` - int
- `depth` - int - Depth of the node
- `children_count` - Number of children

Optional:

- `scope` - restricts what is to be considered a list. You can also setup scope by multiple attributes.

Example:

```go
import (
  "database/sql"
  "github.com/griffinqiu/go-nested-set"
)

// Category
type Category struct {
	ID            int64         `gorm:"PRIMARY_KEY;AUTO_INCREMENT" nestedset:"id"`
	ParentID      sql.NullInt64 `nestedset:"parent_id"`
	UserType      string        `nestedset:"scope"`
	UserID        int64         `nestedset:"scope"`
	Rgt           int           `nestedset:"rgt"`
	Lft           int           `nestedset:"lft"`
	Depth         int           `nestedset:"depth"`
	ChildrenCount int           `nestedset:"children_count"`
	Title         string
}
```

### Move Node

```go
import nestedset "github.com/griffinqiu/go-nested-set"

// nestedset.MoveDirectionLeft
// nestedset.MoveDirectionRight
// nestedset.MoveDirectionInner
nestedset.MoveTo(tx, node, to, nestedset.MoveDirectionLeft)
```

### Get Nodes with tree order

```go
// With scope, limit tree in a scope
tx := db.Model(&Category{}).Where("user_type = ? AND user_id = ?", "User", 100)

// Get all nodes
categories, _ := tx.Order("lft asc").Error

// Get root nodes
categories, _ := tx.Where("parent_id IS NULL").Order("lft asc").Error

// Get childrens
categories, _ := tx.Where("parent_id = ?", parentCategory.ID).Order("lft asc").Error
```

## Testing

```bash
$ createdb nested-set-test
$ go test ./...
```

## License

MIT
