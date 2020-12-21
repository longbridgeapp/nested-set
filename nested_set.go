package nestedset

import (
	"fmt"
	"reflect"

	"github.com/thoas/go-funk"
	"gorm.io/gorm"
)

type MoveDirection int

const (
	MoveDirectionLeft  MoveDirection = 1
	MoveDirectionRight MoveDirection = 2
	MoveDirectionInner MoveDirection = 3
)

type nodeItem struct {
	ID            int64
	ParentId      int64
	Depth         int
	Rgt           int
	Lft           int
	ChildrenCount int
	TableName     string
}

func newNodeItem(db *gorm.DB, source interface{}) (nodeItem, error) {
	err := db.Statement.Parse(source)
	if err != nil {
		return nodeItem{}, fmt.Errorf("Invalid source, must be Gorm Model instance, %v", source)
	}
	item := nodeItem{TableName: db.Statement.Table}
	t := reflect.TypeOf(source)
	tv := reflect.ValueOf(source)
	for i := 0; i < t.NumField(); i++ {
		v := tv.Field(i)
		switch t.Field(i).Tag.Get("nestedset") {
		case "id":
			item.ID = v.Int()
			break
		case "parent_id":
			item.ParentId = v.Int()
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

// MoveTo move node to other node
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
	var newParentId int64
	if direction == MoveDirectionLeft || direction == MoveDirectionRight {
		newParentId = toNode.ParentId
		depthChange = toNode.Depth - targetNode.Depth
		right = toNode.Rgt
		if direction == MoveDirectionLeft {
			right = toNode.Lft - 1
		}
	} else {
		newParentId = toNode.ID
		depthChange = toNode.Depth + 1 - targetNode.Depth
		right = toNode.Lft
	}
	moveToRightOfPosition(db, targetNode, right, depthChange, newParentId)
	return nil
}

func moveToRightOfPosition(db *gorm.DB, targetNode nodeItem, position, depthChange int, newParentId int64) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		oldParentId := targetNode.ParentId
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
			// 向后移需要减去本身的宽度
			moveStep = moveStep - targetWidth
		} else {
			return nil
		}

		err = moveAffected(tx, targetNode.TableName, affectedGte, affectedLte, affectedStep)
		if err != nil {
			return
		}

		err = moveTarget(tx, targetNode.TableName, targetNode.ID, targetIds, moveStep, depthChange, newParentId)
		if err != nil {
			return
		}

		return syncChildrenCount(tx, targetNode.TableName, oldParentId, newParentId)
	})
}

func syncChildrenCount(db *gorm.DB, tableName string, oldParentId, newParentId int64) (err error) {
	ids := []int64{}
	if oldParentId != 0 {
		ids = append(ids, oldParentId)
	}
	if newParentId != 0 {
		ids = append(ids, newParentId)
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

func moveTarget(db *gorm.DB, tableName string, targetId int64, targetIds []int64, step, depthChange int, newParentId int64) (err error) {
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

func moveAffected(db *gorm.DB, tableName string, gte, lte, step int) (err error) {
	sql := fmt.Sprintf(`
UPDATE %s
SET lft=(CASE WHEN lft>=? THEN lft+? ELSE lft END),
	rgt=(CASE WHEN rgt<=? THEN rgt+? ELSE rgt END)
WHERE (lft BETWEEN ? AND ?) OR (rgt BETWEEN ? AND ?);
  `, tableName)
	return db.Exec(sql, gte, step, lte, step, gte, lte, gte, lte).Error
}
