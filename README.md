# go-nested-set

go-nested-set is an implementation of the [Nested set model](https://en.wikipedia.org/wiki/Nested_set_model) for [GORM](https://gorm.io/index.html)

## Usage

### Define the model

```go
import nestedset "github.com/griffinqiu/go-nested-set"

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
