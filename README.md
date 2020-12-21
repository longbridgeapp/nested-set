# Go Nested Set

[![build](https://github.com/griffinqiu/go-nested-set/workflows/build/badge.svg)](https://github.com/griffinqiu/go-nested-set/actions?query=workflow%3Abuild)

Go Nested Set is an implementation of the [Nested set model](https://en.wikipedia.org/wiki/Nested_set_model) for [Gorm](https://gorm.io/index.html).

## Installation

```
go get github.com/griffinqiu/go-nested-set
```

## Usage

### Define the model

```go
import "github.com/griffinqiu/go-nested-set"

// Toc table of contents
type Toc struct {
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
nestedset.MoveTo(gormDB, toc.Node, toc.Node, nestedset.MoveDirectionLeft)
```
