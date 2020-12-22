package nestedset

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReloadData(t *testing.T) {
	reloadCategories()
}

func TestNewNodeItem(t *testing.T) {
	source := Category{
		ID:            123,
		ParentID:      100,
		Depth:         2,
		Rgt:           12,
		Lft:           32,
		ChildrenCount: 10,
	}
	_, node, err := parseNode(gormMock, source)
	assert.NoError(t, err)
	assert.Equal(t, source.ID, node.ID)
	assert.Equal(t, source.ParentID, node.ParentID)
	assert.Equal(t, source.Depth, node.Depth)
	assert.Equal(t, source.Lft, node.Lft)
	assert.Equal(t, source.Rgt, node.Rgt)
	assert.Equal(t, source.ChildrenCount, node.ChildrenCount)
	assert.Equal(t, "categories", node.TableName)
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
		Pid:        101,
		Depth1:     2,
		Right:      10,
		Left:       1,
		NodesCount: 8,
	}
	_, node, err = parseNode(gormMock, specialItem)
	assert.NoError(t, err)
	assert.Equal(t, specialItem.ItemID, node.ID)
	assert.Equal(t, specialItem.Pid, node.ParentID)
	assert.Equal(t, specialItem.Depth1, node.Depth)
	assert.Equal(t, specialItem.Right, node.Rgt)
	assert.Equal(t, specialItem.Left, node.Lft)
	assert.Equal(t, specialItem.NodesCount, node.ChildrenCount)
	assert.Equal(t, "special_items", node.TableName)
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

func TestMoveToRight(t *testing.T) {
	// case 1
	initData()
	db := gormMock.Table("categories")
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

func TestMoveToLeft(t *testing.T) {
	// case 1
	initData()
	db := gormMock.Table("categories")
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
	db := gormMock.Table("categories")
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

func assertNodeEqual(t *testing.T, target Category, left, right, depth, childrenCount int, parentID int64) {
	fmt.Printf("Asserting %s(%d)\n", target.Title, target.ID)
	assert.Equal(t, target.Lft, left)
	assert.Equal(t, target.Rgt, right)
	assert.Equal(t, target.Depth, depth)
	assert.Equal(t, target.ChildrenCount, childrenCount)
	assert.Equal(t, target.ParentID, parentID)
}
