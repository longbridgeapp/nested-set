package nested_set

type Category struct {
	ID       int64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	ParentId int64
	Rgt      int32
	Lft      int32
	Depth    int32
}

func (Category) TableName() string {
	return "course_chapters"
}
