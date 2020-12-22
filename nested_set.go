package nestedset

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

// MoveDirection means where the node is going to be located
type MoveDirection int

const (
	// MoveDirectionLeft : MoveTo(db, a, n, MoveDirectionLeft) => a|n|...
	MoveDirectionLeft MoveDirection = -1

	// MoveDirectionRight : MoveTo(db, a, n, MoveDirectionRight) => ...|n|a|
	MoveDirectionRight MoveDirection = 1

	// MoveDirectionInner : MoveTo(db, a, n, MoveDirectionInner) => [n [...|a]]
	MoveDirectionInner MoveDirection = 0
)

type nodeItem struct {
	ID            int64
	ParentID      int64
	Depth         int
	Rgt           int
	Lft           int
	ChildrenCount int
}

func newNodeItem(db *gorm.DB, source interface{}) (nodeItem, error) {
	err := db.Statement.Parse(source)
	if err != nil {
		return nodeItem{}, fmt.Errorf("Invalid source, must be a valid Gorm Model instance, %v", source)
	}

	item := nodeItem{}
	t := reflect.TypeOf(source)
	v := reflect.ValueOf(source)
	for i := 0; i < t.NumField(); i++ {
		v := v.Field(i)
		switch t.Field(i).Tag.Get("nestedset") {
		case "id":
			item.ID = v.Int()
			break
		case "parent_id":
			item.ParentID = v.Int()
			break
		case "depth":
			item.Depth = int(v.Int())
			break
		case "rgt":
			item.Rgt = int(v.Int())
			break
		case "lft":
			item.Lft = int(v.Int())
			break
		case "children_count":
			item.ChildrenCount = int(v.Int())
			break
		}
	}

	return item, nil
}

// MoveTo move node to a position which is related a target node
func MoveTo(db *gorm.DB, node, to interface{}, direction MoveDirection) error {
	targetNode, err := newNodeItem(db, node)
	if err != nil {
		return err
	}

	toNode, err := newNodeItem(db, to)
	if err != nil {
		return err
	}

	var right, depthChange int
	var newParentID int64
	if direction == MoveDirectionLeft || direction == MoveDirectionRight {
		newParentID = toNode.ParentID
		depthChange = toNode.Depth - targetNode.Depth
		if direction == MoveDirectionLeft {
			right = toNode.Lft - 1
		} else {
			right = toNode.Rgt
		}
	} else {
		newParentID = toNode.ID
		depthChange = toNode.Depth + 1 - targetNode.Depth
		right = toNode.Lft
	}

	return moveToRightOfPosition(db, targetNode, right, depthChange, newParentID)
}

func moveToRightOfPosition(db *gorm.DB, targetNode nodeItem, position, depthChange int, newParentID int64) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		oldParentID := targetNode.ParentID
		targetRight := targetNode.Rgt
		targetLeft := targetNode.Lft
		targetWidth := targetRight - targetLeft + 1

		targetIds := []int64{}
		err = tx.Where("rgt>=? AND rgt <=?", targetLeft, targetRight).Pluck("id", &targetIds).Error
		if err != nil {
			return
		}

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
			// move backwards should minus target covered length/width
			// 向后移需要减去本身的宽度
			moveStep = moveStep - targetWidth
		} else {
			return nil
		}

		err = moveAffected(tx, affectedGte, affectedLte, affectedStep)
		if err != nil {
			return
		}

		err = moveTarget(tx, targetNode.ID, targetIds, moveStep, depthChange, newParentID)
		if err != nil {
			return
		}

		return syncChildrenCount(tx, oldParentID, newParentID)
	})
}

func syncChildrenCount(db *gorm.DB, oldParentID, newParentID int64) (err error) {
	var oldParentCount, newParentCount int64
	if oldParentID != 0 {
		err = db.Where("parent_id=?", oldParentID).Count(&oldParentCount).Error
		if err != nil {
			return
		}
		err = db.Where("id=?", oldParentID).Update("children_count", oldParentCount).Error
		if err != nil {
			return
		}
	}

	if newParentID != 0 {
		err = db.Where("parent_id=?", newParentID).Count(&newParentCount).Error
		if err != nil {
			return
		}
		err = db.Where("id=?", newParentID).Update("children_count", newParentCount).Error
		if err != nil {
			return
		}
	}
	return nil
}

func moveTarget(db *gorm.DB, targetID int64, targetIds []int64, step, depthChange int, newParentID int64) (err error) {
	if len(targetIds) > 0 {
		err = db.Where("id IN (?)", targetIds).
			Updates(map[string]interface{}{
				"lft":   gorm.Expr("lft+?", step),
				"rgt":   gorm.Expr("rgt+?", step),
				"depth": gorm.Expr("depth+?", depthChange),
			}).Error
		if err != nil {
			return
		}
	}
	return db.Where("id=?", targetID).Update("parent_id", newParentID).Error
}

func moveAffected(db *gorm.DB, gte, lte, step int) (err error) {
	return db.Where("(lft BETWEEN ? AND ?) OR (rgt BETWEEN ? AND ?)", gte, lte, gte, lte).
		Updates(map[string]interface{}{
			"lft": gorm.Expr("(CASE WHEN lft>=? THEN lft+? ELSE lft END)", gte, step),
			"rgt": gorm.Expr("(CASE WHEN rgt<=? THEN rgt+? ELSE rgt END)", lte, step),
		}).Error
}
