package nestedset

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/clause"
)

func TestReloadData(t *testing.T) {
	reloadCategories()
}

func TestNewNodeItem(t *testing.T) {
	source := Category{
		ID:            123,
		ParentID:      sql.NullInt64{Valid: true, Int64: 100},
		Depth:         2,
		Rgt:           12,
		Lft:           32,
		UserType:      "User",
		UserID:        1000,
		ChildrenCount: 10,
	}
	tx, node, err := parseNode(db, source)
	assert.NoError(t, err)
	assert.Equal(t, source.ID, node.ID)
	assert.Equal(t, source.ParentID, node.ParentID)
	assert.Equal(t, source.Depth, node.Depth)
	assert.Equal(t, source.Lft, node.Lft)
	assert.Equal(t, source.Rgt, node.Rgt)
	assert.Equal(t, source.ChildrenCount, node.ChildrenCount)
	assert.Equal(t, "categories", node.TableName)
	stmt := tx.Statement
	stmt.Build(clause.Where{}.Name())
	assert.Equal(t, "WHERE user_id = $1 AND user_type = $2", stmt.SQL.String())

	tx, node, err = parseNode(db, &source)
	assert.NoError(t, err)
	assert.Equal(t, source.ID, node.ID)
	assert.Equal(t, source.ParentID, node.ParentID)
	assert.Equal(t, source.Depth, node.Depth)
	assert.Equal(t, source.Lft, node.Lft)
	assert.Equal(t, source.Rgt, node.Rgt)
	assert.Equal(t, source.ChildrenCount, node.ChildrenCount)
	assert.Equal(t, "categories", node.TableName)
	stmt = tx.Statement
	stmt.Build(clause.Where{}.Name())
	assert.Equal(t, "WHERE user_id = $1 AND user_type = $2", stmt.SQL.String())

	dbNames := node.DbNames
	assert.Equal(t, "id", dbNames["id"])
	assert.Equal(t, "parent_id", dbNames["parent_id"])
	assert.Equal(t, "depth", dbNames["depth"])
	assert.Equal(t, "rgt", dbNames["rgt"])
	assert.Equal(t, "lft", dbNames["lft"])
	assert.Equal(t, "children_count", dbNames["children_count"])

	// Test for difference column names
	specialItem := SpecialItem{
		ItemID:     100,
		Pid:        sql.NullInt64{Valid: true, Int64: 101},
		Depth1:     2,
		Right:      10,
		Left:       1,
		NodesCount: 8,
	}
	tx, node, err = parseNode(db, specialItem)
	assert.NoError(t, err)
	assert.Equal(t, specialItem.ItemID, node.ID)
	assert.Equal(t, specialItem.Pid, node.ParentID)
	assert.Equal(t, specialItem.Depth1, node.Depth)
	assert.Equal(t, specialItem.Right, node.Rgt)
	assert.Equal(t, specialItem.Left, node.Lft)
	assert.Equal(t, specialItem.NodesCount, node.ChildrenCount)
	assert.Equal(t, "special_items", node.TableName)

	stmt = tx.Statement
	stmt.Build(clause.Where{}.Name())
	assert.Equal(t, "", stmt.SQL.String())

	dbNames = node.DbNames
	assert.Equal(t, "item_id", dbNames["id"])
	assert.Equal(t, "pid", dbNames["parent_id"])
	assert.Equal(t, "depth1", dbNames["depth"])
	assert.Equal(t, "right", dbNames["rgt"])
	assert.Equal(t, "left", dbNames["lft"])
	assert.Equal(t, "nodes_count", dbNames["children_count"])

	// formatSQL test
	assert.Equal(t, "item_id = ? AND left > right AND pid = ?, nodes_count = 1, depth1 = 0", formatSQL(":id = ? AND :lft > :rgt AND :parent_id = ?, :children_count = 1, :depth = 0", node))
}

func TestCreateSource(t *testing.T) {
	initData()

	c1 := Category{Title: "c1s"}
	var cNil *Category
	Create(db, &c1, cNil)
	assert.Equal(t, c1.Lft, 1)
	assert.Equal(t, c1.Rgt, 2)
	assert.Equal(t, c1.Depth, 0)

	cp := Category{Title: "cps"}
	Create(db, &cp, nil)
	assert.Equal(t, cp.Lft, 3)
	assert.Equal(t, cp.Rgt, 4)

	c2 := Category{Title: "c2s", UserType: "ux"}
	Create(db, &c2, nil)
	assert.Equal(t, c2.Lft, 1)
	assert.Equal(t, c2.Rgt, 2)

	c3 := Category{Title: "c3s", UserType: "ux"}
	Create(db, &c3, nil)
	assert.Equal(t, c3.Lft, 3)
	assert.Equal(t, c3.Rgt, 4)

	c4 := Category{Title: "c4s", UserType: "ux"}
	Create(db, &c4, &c2)
	assert.Equal(t, c4.Lft, 2)
	assert.Equal(t, c4.Rgt, 3)
	assert.Equal(t, c4.Depth, 1)

	// after insert a new node into c2
	db.Find(&c3)
	db.Find(&c2)
	assert.Equal(t, c3.Lft, 5)
	assert.Equal(t, c3.Rgt, 6)
	assert.Equal(t, c2.ChildrenCount, 1)
}

func TestDeleteSource(t *testing.T) {
	initData()

	c1 := Category{Title: "c1s"}
	Create(db, &c1, nil)

	cp := Category{Title: "cp"}
	Create(db, &cp, c1)

	c2 := Category{Title: "c2s"}
	Create(db, &c2, nil)

	db.First(&c1)
	Delete(db, &c1)

	db.Model(&c2).First(&c2)
	assert.Equal(t, c2.Lft, 1)
	assert.Equal(t, c2.Rgt, 2)
}

func TestMoveToRight(t *testing.T) {
	// case 1
	initData()
	MoveTo(db, dresses, jackets, MoveDirectionRight)
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, 0)
	assertNodeEqual(t, mens, 2, 15, 1, 1, clothing.ID)
	assertNodeEqual(t, suits, 3, 14, 2, 3, mens.ID)
	assertNodeEqual(t, slacks, 4, 5, 3, 0, suits.ID)
	assertNodeEqual(t, jackets, 6, 7, 3, 0, suits.ID)
	assertNodeEqual(t, dresses, 8, 13, 3, 2, suits.ID)
	assertNodeEqual(t, eveningGowns, 9, 10, 4, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 11, 12, 4, 0, dresses.ID)
	assertNodeEqual(t, womens, 16, 21, 1, 2, clothing.ID)
	assertNodeEqual(t, skirts, 17, 18, 2, 0, womens.ID)
	assertNodeEqual(t, blouses, 19, 20, 2, 0, womens.ID)

	// case 2
	initData()
	MoveTo(db, suits, blouses, MoveDirectionRight)
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, 0)
	assertNodeEqual(t, mens, 2, 3, 1, 0, clothing.ID)
	assertNodeEqual(t, womens, 4, 21, 1, 4, clothing.ID)
	assertNodeEqual(t, dresses, 5, 10, 2, 2, womens.ID)
	assertNodeEqual(t, eveningGowns, 6, 7, 3, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 8, 9, 3, 0, dresses.ID)
	assertNodeEqual(t, skirts, 11, 12, 2, 0, womens.ID)
	assertNodeEqual(t, blouses, 13, 14, 2, 0, womens.ID)
	assertNodeEqual(t, suits, 15, 20, 2, 2, womens.ID)
	assertNodeEqual(t, slacks, 16, 17, 3, 0, suits.ID)
	assertNodeEqual(t, jackets, 18, 19, 3, 0, suits.ID)
}

func TestRebuild(t *testing.T) {
	initData()
	affectedCount, err := Rebuild(db, clothing, true)
	assert.NoError(t, err)
	assert.Equal(t, 0, affectedCount)
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, 0)
	assertNodeEqual(t, mens, 2, 9, 1, 1, clothing.ID)
	assertNodeEqual(t, suits, 3, 8, 2, 2, mens.ID)
	assertNodeEqual(t, slacks, 4, 5, 3, 0, suits.ID)
	assertNodeEqual(t, jackets, 6, 7, 3, 0, suits.ID)
	assertNodeEqual(t, womens, 10, 21, 1, 3, clothing.ID)
	assertNodeEqual(t, dresses, 11, 16, 2, 2, womens.ID)
	assertNodeEqual(t, eveningGowns, 12, 13, 3, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 14, 15, 3, 0, dresses.ID)
	assertNodeEqual(t, skirts, 17, 18, 2, 0, womens.ID)
	assertNodeEqual(t, blouses, 19, 20, 2, 0, womens.ID)

	sunDresses.Rgt = 123
	sunDresses.Lft = 12
	sunDresses.Depth = 1
	sunDresses.ChildrenCount = 100
	err = db.Updates(&sunDresses).Error
	assert.NoError(t, err)
	reloadCategories()
	assertNodeEqual(t, sunDresses, 12, 123, 1, 100, dresses.ID)

	affectedCount, err = Rebuild(db, clothing, true)
	assert.NoError(t, err)
	assert.Equal(t, 2, affectedCount)
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, 0)
	assertNodeEqual(t, mens, 2, 9, 1, 1, clothing.ID)
	assertNodeEqual(t, suits, 3, 8, 2, 2, mens.ID)
	assertNodeEqual(t, slacks, 4, 5, 3, 0, suits.ID)
	assertNodeEqual(t, jackets, 6, 7, 3, 0, suits.ID)
	assertNodeEqual(t, womens, 10, 21, 1, 3, clothing.ID)
	assertNodeEqual(t, dresses, 11, 16, 2, 2, womens.ID)
	assertNodeEqual(t, eveningGowns, 14, 15, 3, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 12, 13, 3, 0, dresses.ID)
	assertNodeEqual(t, skirts, 17, 18, 2, 0, womens.ID)
	assertNodeEqual(t, blouses, 19, 20, 2, 0, womens.ID)

	affectedCount, err = Rebuild(db, clothing, true)
	assert.NoError(t, err)
	assert.Equal(t, 0, affectedCount)
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, 0)
	assertNodeEqual(t, mens, 2, 9, 1, 1, clothing.ID)
	assertNodeEqual(t, suits, 3, 8, 2, 2, mens.ID)
	assertNodeEqual(t, slacks, 4, 5, 3, 0, suits.ID)
	assertNodeEqual(t, jackets, 6, 7, 3, 0, suits.ID)
	assertNodeEqual(t, womens, 10, 21, 1, 3, clothing.ID)
	assertNodeEqual(t, dresses, 11, 16, 2, 2, womens.ID)
	assertNodeEqual(t, eveningGowns, 14, 15, 3, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 12, 13, 3, 0, dresses.ID)
	assertNodeEqual(t, skirts, 17, 18, 2, 0, womens.ID)
	assertNodeEqual(t, blouses, 19, 20, 2, 0, womens.ID)

	hat := *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Hat",
		"ParentID": sql.NullInt64{Valid: false},
	}).(*Category)

	affectedCount, err = Rebuild(db, clothing, false)
	assert.NoError(t, err)
	assert.Equal(t, 1, affectedCount)

	affectedCount, err = Rebuild(db, clothing, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, affectedCount)
	reloadCategories()
	hat, _ = findNode(db, hat.ID)

	assertNodeEqual(t, clothing, 1, 22, 0, 2, 0)
	assertNodeEqual(t, mens, 2, 9, 1, 1, clothing.ID)
	assertNodeEqual(t, suits, 3, 8, 2, 2, mens.ID)
	assertNodeEqual(t, slacks, 4, 5, 3, 0, suits.ID)
	assertNodeEqual(t, jackets, 6, 7, 3, 0, suits.ID)
	assertNodeEqual(t, womens, 10, 21, 1, 3, clothing.ID)
	assertNodeEqual(t, dresses, 11, 16, 2, 2, womens.ID)
	assertNodeEqual(t, eveningGowns, 14, 15, 3, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 12, 13, 3, 0, dresses.ID)
	assertNodeEqual(t, skirts, 17, 18, 2, 0, womens.ID)
	assertNodeEqual(t, blouses, 19, 20, 2, 0, womens.ID)
	assertNodeEqual(t, hat, 23, 24, 0, 0, 0)

	jacksClothing := *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Jack's Clothing",
		"ParentID": sql.NullInt64{Valid: false},
		"UserType": "User",
		"UserID":   8686,
	}).(*Category)
	jacksSuits := *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Jack's Suits",
		"ParentID": sql.NullInt64{Valid: true, Int64: jacksClothing.ID},
		"UserType": "User",
		"UserID":   8686,
	}).(*Category)
	jacksHat := *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Jack's Hat",
		"UserType": "User",
		"UserID":   8686,
		"ParentID": sql.NullInt64{Valid: false},
	}).(*Category)
	jacksSlacks := *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Jack's Slacks",
		"ParentID": sql.NullInt64{Valid: true, Int64: jacksClothing.ID},
		"UserType": "User",
		"UserID":   8686,
	}).(*Category)

	lilysHat := *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Lily's Hat",
		"UserType": "User",
		"UserID":   6666,
		"ParentID": sql.NullInt64{Valid: false},
	}).(*Category)
	lilysClothing := *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Lily's Clothing",
		"ParentID": sql.NullInt64{Valid: false},
		"UserType": "User",
		"UserID":   6666,
	}).(*Category)
	lilysDresses := *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Lily's Dresses",
		"ParentID": sql.NullInt64{Valid: true, Int64: lilysClothing.ID},
		"UserType": "User",
		"UserID":   6666,
	}).(*Category)

	affectedCount, err = Rebuild(db, jacksSuits, true)
	assert.NoError(t, err)
	assert.Equal(t, 4, affectedCount)
	affectedCount, err = Rebuild(db, lilysHat, true)
	assert.NoError(t, err)
	assert.Equal(t, 3, affectedCount)
	jacksClothing, _ = findNode(db, jacksClothing.ID)
	jacksSuits, _ = findNode(db, jacksSuits.ID)
	jacksSlacks, _ = findNode(db, jacksSlacks.ID)
	jacksHat, _ = findNode(db, jacksHat.ID)
	lilysHat, _ = findNode(db, lilysHat.ID)
	lilysClothing, _ = findNode(db, lilysClothing.ID)
	lilysDresses, _ = findNode(db, lilysDresses.ID)

	assertNodeEqual(t, jacksClothing, 1, 6, 0, 2, 0)
	assertNodeEqual(t, jacksSuits, 2, 3, 1, 0, jacksClothing.ID)
	assertNodeEqual(t, jacksSlacks, 4, 5, 1, 0, jacksClothing.ID)
	assertNodeEqual(t, jacksHat, 7, 8, 0, 0, 0)
	assertNodeEqual(t, lilysHat, 1, 2, 0, 0, 0)
	assertNodeEqual(t, lilysClothing, 3, 6, 0, 1, 0)
	assertNodeEqual(t, lilysDresses, 4, 5, 1, 0, lilysClothing.ID)
}

func TestMoveToLeft(t *testing.T) {
	// case 1
	initData()
	MoveTo(db, dresses, jackets, MoveDirectionLeft)
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, 0)
	assertNodeEqual(t, mens, 2, 15, 1, 1, clothing.ID)
	assertNodeEqual(t, suits, 3, 14, 2, 3, mens.ID)
	assertNodeEqual(t, slacks, 4, 5, 3, 0, suits.ID)
	assertNodeEqual(t, dresses, 6, 11, 3, 2, suits.ID)
	assertNodeEqual(t, eveningGowns, 7, 8, 4, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 9, 10, 4, 0, dresses.ID)
	assertNodeEqual(t, jackets, 12, 13, 3, 0, suits.ID)
	assertNodeEqual(t, womens, 16, 21, 1, 2, clothing.ID)
	assertNodeEqual(t, skirts, 17, 18, 2, 0, womens.ID)
	assertNodeEqual(t, blouses, 19, 20, 2, 0, womens.ID)

	// case 2
	initData()
	MoveTo(db, suits, blouses, MoveDirectionLeft)
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, 0)
	assertNodeEqual(t, mens, 2, 3, 1, 0, clothing.ID)
	assertNodeEqual(t, womens, 4, 21, 1, 4, clothing.ID)
	assertNodeEqual(t, dresses, 5, 10, 2, 2, womens.ID)
	assertNodeEqual(t, eveningGowns, 6, 7, 3, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 8, 9, 3, 0, dresses.ID)
	assertNodeEqual(t, skirts, 11, 12, 2, 0, womens.ID)
	assertNodeEqual(t, suits, 13, 18, 2, 2, womens.ID)
	assertNodeEqual(t, slacks, 14, 15, 3, 0, suits.ID)
	assertNodeEqual(t, jackets, 16, 17, 3, 0, suits.ID)
	assertNodeEqual(t, blouses, 19, 20, 2, 0, womens.ID)
}

func TestMoveToInner(t *testing.T) {
	// case 1
	initData()
	MoveTo(db, mens, blouses, MoveDirectionInner)
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 1, 0)
	assertNodeEqual(t, womens, 2, 21, 1, 3, clothing.ID)
	assertNodeEqual(t, dresses, 3, 8, 2, 2, womens.ID)
	assertNodeEqual(t, eveningGowns, 4, 5, 3, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 6, 7, 3, 0, dresses.ID)
	assertNodeEqual(t, skirts, 9, 10, 2, 0, womens.ID)
	assertNodeEqual(t, blouses, 11, 20, 2, 1, womens.ID)
	assertNodeEqual(t, mens, 12, 19, 3, 1, blouses.ID)
	assertNodeEqual(t, suits, 13, 18, 4, 2, mens.ID)
	assertNodeEqual(t, slacks, 14, 15, 5, 0, suits.ID)
	assertNodeEqual(t, jackets, 16, 17, 5, 0, suits.ID)

	// case 2
	initData()
	MoveTo(db, skirts, slacks, MoveDirectionInner)
	reloadCategories()

	assertNodeEqual(t, clothing, 1, 22, 0, 2, 0)
	assertNodeEqual(t, mens, 2, 11, 1, 1, clothing.ID)
	assertNodeEqual(t, suits, 3, 10, 2, 2, mens.ID)
	assertNodeEqual(t, slacks, 4, 7, 3, 1, suits.ID)
	assertNodeEqual(t, skirts, 5, 6, 4, 0, slacks.ID)
	assertNodeEqual(t, jackets, 8, 9, 3, 0, suits.ID)
	assertNodeEqual(t, womens, 12, 21, 1, 2, clothing.ID)
	assertNodeEqual(t, dresses, 13, 18, 2, 2, womens.ID)
	assertNodeEqual(t, eveningGowns, 14, 15, 3, 0, dresses.ID)
	assertNodeEqual(t, sunDresses, 16, 17, 3, 0, dresses.ID)
	assertNodeEqual(t, blouses, 19, 20, 2, 0, womens.ID)
}

func TestMoveIsInvalid(t *testing.T) {
	initData()
	err := MoveTo(db, womens, dresses, MoveDirectionInner)
	assert.NotEmpty(t, err)
	reloadCategories()
	assertNodeEqual(t, womens, 10, 21, 1, 3, clothing.ID)

	err = MoveTo(db, womens, dresses, MoveDirectionLeft)
	assert.NotEmpty(t, err)
	reloadCategories()
	assertNodeEqual(t, womens, 10, 21, 1, 3, clothing.ID)

	err = MoveTo(db, womens, dresses, MoveDirectionRight)
	assert.NotEmpty(t, err)
	reloadCategories()
	assertNodeEqual(t, womens, 10, 21, 1, 3, clothing.ID)
}

func assertNodeEqual(t *testing.T, target Category, left, right, depth, childrenCount int, parentID int64) {
	nullInt64ParentID := sql.NullInt64{Valid: false}
	if parentID > 0 {
		nullInt64ParentID = sql.NullInt64{Valid: true, Int64: parentID}
	}
	assert.Equal(t, left, target.Lft)
	assert.Equal(t, right, target.Rgt)
	assert.Equal(t, depth, target.Depth)
	assert.Equal(t, childrenCount, target.ChildrenCount)
	assert.Equal(t, nullInt64ParentID, target.ParentID)
}
