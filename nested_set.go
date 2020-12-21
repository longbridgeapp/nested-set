package nestedset

import (
	"fmt"
	"reflect"

	"github.com/thoas/go-funk"
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
	TableName     string
}

func newNodeItem(db *gorm.DB, source interface{}) (nodeItem, error) {
	err := db.Statement.Parse(source)
	if err != nil {
		return nodeItem{}, fmt.Errorf("Invalid source, must be a valid Gorm Model instance, %v", source)
	}

	item := nodeItem{TableName: db.Statement.Table}
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
func MoveTo(db *gorm.DB, node interface{}, to interface{}, direction MoveDirection) error {
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

	moveToRightOfPosition(db, targetNode, right, depthChange, newParentID)
	return nil
}

func moveToRightOfPosition(db *gorm.DB, targetNode nodeItem, position, depthChange int, newParentID int64) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		oldParentID := targetNode.ParentID
		targetRight := targetNode.Rgt
		targetLeft := targetNode.Lft
		targetWidth := targetRight - targetLeft + 1

		targetIds := []int64{}
		err = tx.Table(targetNode.TableName).Where("rgt>=? AND rgt <=?", targetLeft, targetRight).Pluck("id", &targetIds).Error
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

		err = moveAffected(tx, targetNode.TableName, affectedGte, affectedLte, affectedStep)
		if err != nil {
			return
		}

		err = moveTarget(tx, targetNode.TableName, targetNode.ID, targetIds, moveStep, depthChange, newParentID)
		if err != nil {
			return
		}

		return syncChildrenCount(tx, targetNode.TableName, oldParentID, newParentID)
	})
}

func syncChildrenCount(db *gorm.DB, tableName string, oldParentID, newParentID int64) (err error) {
	ids := []int64{}
	if oldParentID != 0 {
		ids = append(ids, oldParentID)
	}
	if newParentID != 0 {
		ids = append(ids, newParentID)
	}
	if len(ids) == 0 {
		return nil
	}

	ids = funk.UniqInt64(ids)

	sql := fmt.Sprintf(`
UPDATE %s as a
SET children_count=(SELECT COUNT(1) FROM %s AS b WHERE a.id=b.parent_id)
WHERE a.id IN (?)
	`, tableName, tableName)

	return db.Exec(sql, ids).Error
}

func moveTarget(db *gorm.DB, tableName string, targetID int64, targetIds []int64, step, depthChange int, newParentID int64) (err error) {
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
	return db.Exec(fmt.Sprintf("UPDATE %s SET parent_id=? WHERE id=?", tableName), newParentID, targetID).Error
}

func moveAffected(db *gorm.DB, tableName string, gte, lte, step int) (err error) {
	sql := fmt.Sprintf(`
UPDATE %s
SET lft=(CASE WHEN lft>=? THEN lft+? ELSE lft END),
	rgt=(CASE WHEN rgt<=? THEN rgt+? ELSE rgt END)
WHERE (lft BETWEEN ? AND ?) OR (rgt BETWEEN ? AND ?);
  `, tableName)
	return db.Exec(sql, gte, step, lte, step, gte, lte, gte, lte).Error
}
