package nestedset

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	ParentID      sql.NullInt64
	Depth         int
	Rgt           int
	Lft           int
	ChildrenCount int
	TableName     string
	DbNames       map[string]string
}

func parseNode(db *gorm.DB, source interface{}) (tx *gorm.DB, item nodeItem, err error) {
	tx = db
	stmt := &gorm.Statement{
		DB:       tx,
		ConnPool: tx.ConnPool,
		Context:  context.Background(),
		Clauses:  map[string]clause.Clause{},
	}

	err = stmt.Parse(source)
	if err != nil {
		err = fmt.Errorf("Invalid source, must be a valid Gorm Model instance, %v", source)
		return
	}

	item = nodeItem{TableName: stmt.Table, DbNames: map[string]string{}}
	sourceValue := reflect.Indirect(reflect.ValueOf(source))
	sourceType := sourceValue.Type()
	for i := 0; i < sourceType.NumField(); i++ {
		t := sourceType.Field(i)
		v := sourceValue.Field(i)

		schemaField := stmt.Schema.LookUpField(t.Name)
		dbName := schemaField.DBName

		switch t.Tag.Get("nestedset") {
		case "id":
			item.ID = v.Int()
			item.DbNames["id"] = dbName
			break
		case "parent_id":
			item.ParentID = v.Interface().(sql.NullInt64)
			item.DbNames["parent_id"] = dbName
			break
		case "depth":
			item.Depth = int(v.Int())
			item.DbNames["depth"] = dbName
			break
		case "rgt":
			item.Rgt = int(v.Int())
			item.DbNames["rgt"] = dbName
			break
		case "lft":
			item.Lft = int(v.Int())
			item.DbNames["lft"] = dbName
			break
		case "children_count":
			item.ChildrenCount = int(v.Int())
			item.DbNames["children_count"] = dbName
			break
		case "scope":
			rawVal, _ := schemaField.ValueOf(sourceValue)
			tx = tx.Where(dbName+" = ?", rawVal)
			break
		}
	}

	return
}

// MoveTo move node to a position which is related a target node
func MoveTo(db *gorm.DB, node, to interface{}, direction MoveDirection) error {
	_, targetNode, err := parseNode(db, node)
	if err != nil {
		return err
	}

	_, toNode, err := parseNode(db, to)
	if err != nil {
		return err
	}

	tx := db.Table(targetNode.TableName)

	var right, depthChange int
	var newParentID sql.NullInt64
	if direction == MoveDirectionLeft || direction == MoveDirectionRight {
		newParentID = toNode.ParentID
		depthChange = toNode.Depth - targetNode.Depth
		if direction == MoveDirectionLeft {
			right = toNode.Lft - 1
		} else {
			right = toNode.Rgt
		}
	} else {
		newParentID = sql.NullInt64{Int64: toNode.ID, Valid: true}
		depthChange = toNode.Depth + 1 - targetNode.Depth
		right = toNode.Lft
	}

	return moveToRightOfPosition(tx, targetNode, right, depthChange, newParentID)
}

func moveToRightOfPosition(tx *gorm.DB, targetNode nodeItem, position, depthChange int, newParentID sql.NullInt64) error {
	return tx.Transaction(func(tx *gorm.DB) (err error) {
		oldParentID := targetNode.ParentID
		targetRight := targetNode.Rgt
		targetLeft := targetNode.Lft
		targetWidth := targetRight - targetLeft + 1

		targetIds := []int64{}
		err = tx.Where(formatSQL(":lft >= ? AND :rgt <= ?", targetNode), targetLeft, targetRight).Pluck("id", &targetIds).Error
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
			moveStep = moveStep - targetWidth
		} else {
			return nil
		}

		err = moveAffected(tx, targetNode, affectedGte, affectedLte, affectedStep)
		if err != nil {
			return
		}

		err = moveTarget(tx, targetNode, targetNode.ID, targetIds, moveStep, depthChange, newParentID)
		if err != nil {
			return
		}

		return syncChildrenCount(tx, targetNode, oldParentID, newParentID)
	})
}

func syncChildrenCount(tx *gorm.DB, targetNode nodeItem, oldParentID, newParentID sql.NullInt64) (err error) {
	var oldParentCount, newParentCount int64
	if oldParentID.Valid {
		err = tx.Where(formatSQL(":parent_id = ?", targetNode), oldParentID).Count(&oldParentCount).Error
		if err != nil {
			return
		}
		err = tx.Where(formatSQL(":id = ?", targetNode), oldParentID).Update(targetNode.DbNames["children_count"], oldParentCount).Error
		if err != nil {
			return
		}
	}

	if newParentID.Valid {
		err = tx.Where(formatSQL(":parent_id = ?", targetNode), newParentID).Count(&newParentCount).Error
		if err != nil {
			return
		}
		err = tx.Where(formatSQL(":id = ?", targetNode), newParentID).Update(targetNode.DbNames["children_count"], newParentCount).Error
		if err != nil {
			return
		}
	}
	return nil
}

func moveTarget(tx *gorm.DB, targetNode nodeItem, targetID int64, targetIds []int64, step, depthChange int, newParentID sql.NullInt64) (err error) {
	dbNames := targetNode.DbNames

	if len(targetIds) > 0 {
		err = tx.Where(formatSQL(":id IN (?)", targetNode), targetIds).
			Updates(map[string]interface{}{
				dbNames["lft"]:   gorm.Expr(formatSQL(":lft + ?", targetNode), step),
				dbNames["rgt"]:   gorm.Expr(formatSQL(":rgt + ?", targetNode), step),
				dbNames["depth"]: gorm.Expr(formatSQL(":depth + ?", targetNode), depthChange),
			}).Error
		if err != nil {
			return
		}
	}
	return tx.Where(formatSQL(":id = ?", targetNode), targetID).Update(dbNames["parent_id"], newParentID).Error
}

func moveAffected(tx *gorm.DB, targetNode nodeItem, gte, lte, step int) (err error) {
	dbNames := targetNode.DbNames

	return tx.Where(formatSQL("(:lft BETWEEN ? AND ?) OR (:rgt BETWEEN ? AND ?)", targetNode), gte, lte, gte, lte).
		Updates(map[string]interface{}{
			dbNames["lft"]: gorm.Expr(formatSQL("(CASE WHEN :lft >= ? THEN :lft + ? ELSE :lft END)", targetNode), gte, step),
			dbNames["rgt"]: gorm.Expr(formatSQL("(CASE WHEN :rgt <= ? THEN :rgt + ? ELSE :rgt END)", targetNode), lte, step),
		}).Error
}

func formatSQL(placeHolderSQL string, node nodeItem) (out string) {
	out = placeHolderSQL

	out = strings.ReplaceAll(out, ":table_name", node.TableName)
	for k, v := range node.DbNames {
		out = strings.Replace(out, ":"+k, v, -1)
	}

	return
}
