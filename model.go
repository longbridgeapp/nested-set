package nestedset

import "database/sql"

type Node struct {
	ID            int64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Title         string
	ParentId      sql.NullInt64
	Rgt           int
	Lft           int
	Depth         int
	ChildrenCount int
}

func (Node) TableName() string {
	return "course_chapters"
}
