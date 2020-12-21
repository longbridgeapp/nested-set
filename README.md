# Go Nested Set

[![build](https://github.com/griffinqiu/go-nested-set/workflows/build/badge.svg)](https://github.com/griffinqiu/go-nested-set/actions?query=workflow%3Abuild)

Go Nested Set is an implementation of the [Nested set model](https://en.wikipedia.org/wiki/Nested_set_model) for [Gorm](https://gorm.io/index.html).

This is a Go version of the [awesome_nested_set](https://github.com/collectiveidea/awesome_nested_set).

## Database struct

```go
type Category struct {
	ID            int64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Name          string
	ParentId      sql.NullInt64
	Rgt           int
	Lft           int
	Depth         int
	ChildrenCount int
}
```

## Installation

```
go get github.com/griffinqiu/go-nested-set
```

## Usage

### Define the model

```go
import "github.com/griffinqiu/go-nested-set"

// Toc table of contents
type Category struct {
	nestedset.Node
	Name string
	Status int
}
```

### Move Node

```go
import nestedset "github.com/griffinqiu/go-nested-set"

// nestedset.MoveDirectionLeft
// nestedset.MoveDirectionRight
// nestedset.MoveDirectionInner
nestedset.MoveTo(gormDB, category.Node, category.Node, nestedset.MoveDirectionLeft)
```
