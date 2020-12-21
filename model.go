package nested_set

import "database/sql"

type Category struct {
	ID            int64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Title         string
	ParentId      sql.NullInt64
	Rgt           int
	Lft           int
	Depth         int
	ChildrenCount int
}

func (Category) TableName() string {
	return "course_chapters"
}
