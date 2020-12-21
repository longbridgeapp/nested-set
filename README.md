# go-nested-set

go-nested-set is an implementation of the [Nested set model](https://en.wikipedia.org/wiki/Nested_set_model) for [GORM](https://gorm.io/index.html)


## Usage

```go
type CourseChapter struct {
	nestedset.Category
	Status      int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}


nestedset.MoveTo(gormDB, chapter.Category, toChapter.Category, nestedset.MoveDirectionLeft)
```
