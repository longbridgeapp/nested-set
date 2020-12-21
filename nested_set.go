package nested_set

import (
	"database/sql"
	"fmt"

	"github.com/thoas/go-funk"
	"gorm.io/gorm"
)

type MoveDirection int

const (
	MoveDirectionLeft  MoveDirection = 1
	MoveDirectionRight MoveDirection = 2
	MoveDirectionInner MoveDirection = 3
)

func MoveTo(db *gorm.DB, target Node, to Node, direction MoveDirection) error {
	var right, depthChange int
	var newParentId sql.NullInt64
	if direction == MoveDirectionLeft || direction == MoveDirectionRight {
		newParentId = to.ParentId
		depthChange = to.Depth - target.Depth
		right = to.Rgt
		if direction == MoveDirectionLeft {
			right = to.Lft - 1
		}
	} else {
		newParentId = sql.NullInt64{Int64: to.ID, Valid: true}
		depthChange = to.Depth + 1 - target.Depth
		right = to.Lft
	}
	moveToRightOfPosition(db, target, right, depthChange, newParentId)
	return nil
}

func moveToRightOfPosition(db *gorm.DB, target Node, position, depthChange int, newParentId sql.NullInt64) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		oldParentId := target.ParentId
		targetRight := target.Rgt
		targetLeft := target.Lft
		targetWidth := targetRight - targetLeft + 1

		targets, err := findCategories(tx, targetLeft, targetRight)
		if err != nil {
			return
		}

		targetIds := funk.Map(targets, func(c Node) int64 {
			return c.ID
		}).([]int64)

		var moveStep, affectedStep, affectedGte, affectedLte int
		moveStep = position - targetLeft + 1
		if moveStep < 0 {
			affectedGte = position + 1
			affectedLte = targetLeft - 1
			affectedStep = targetWidth
		} else if moveStep > 0 {
			affectedGte = targetRight + 1
			affectedLte = position
			affectedStep = targetWidth * -1
			// 向后移需要减去本身的宽度
			moveStep = moveStep - targetWidth
		} else {
			return nil
		}

		err = moveAffected(tx, affectedGte, affectedLte, affectedStep)
		if err != nil {
			return
		}

		err = moveTarget(tx, target.ID, targetIds, moveStep, depthChange, newParentId)
		if err != nil {
			return
		}

		return syncChildrenCount(tx, oldParentId, newParentId)
	})
}

func syncChildrenCount(db *gorm.DB, oldParentId, newParentId sql.NullInt64) (err error) {
	ids := []int64{}
	if oldParentId.Valid {
		ids = append(ids, oldParentId.Int64)
	}
	if newParentId.Valid {
		ids = append(ids, newParentId.Int64)
	}
	if len(ids) == 0 {
		return nil
	}

	tableName := Node{}.TableName()
	sql := fmt.Sprintf(`
UPDATE %s as a
SET children_count=(SELECT COUNT(1) FROM course_chapters AS b WHERE a.id=b.parent_id)
WHERE a.id IN (?)
  `, tableName)

	return db.Exec(sql, ids).Error
}

func moveTarget(db *gorm.DB, targetId int64, targetIds []int64, step, depthChange int, newParentId sql.NullInt64) (err error) {
	tableName := Node{}.TableName()
	sql := fmt.Sprintf(`
UPDATE %s
SET lft=lft+?,
	rgt=rgt+?,
	depth=depth+?
WHERE id IN (?);
  `, tableName)
	err = db.Exec(sql, step, step, depthChange, targetIds).Error
	if err != nil {
		return
	}
	return db.Exec(fmt.Sprintf("UPDATE %s SET parent_id=? WHERE id=?", tableName), newParentId, targetId).Error
}

func moveAffected(db *gorm.DB, gte, lte, step int) (err error) {
	tableName := Node{}.TableName()
	sql := fmt.Sprintf(`
UPDATE %s
SET lft=(CASE WHEN lft>=? THEN lft+? ELSE lft END),
	rgt=(CASE WHEN rgt<=? THEN rgt+? ELSE rgt END)
WHERE (lft BETWEEN ? AND ?) OR (rgt BETWEEN ? AND ?);
  `, tableName)
	return db.Exec(sql, gte, step, lte, step, gte, lte, gte, lte).Error
}

func findCategories(query *gorm.DB, left, right int) (categories []Node, err error) {
	err = query.Where("rgt>=? AND rgt <=?", left, right).Find(&categories).Error
	return
}

func findNode(query *gorm.DB, id int64) (Node Node, err error) {
	err = query.Where("id=?", id).Find(&Node).Error
	return
}
