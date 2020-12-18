package nested_set

type Category struct {
	ID       int64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Title    string
	ParentId int64
	Rgt      int
	Lft      int
	Depth    int
}

func (Category) TableName() string {
	return "course_chapters"
}
