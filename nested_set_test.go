package nested_set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReloadData(t *testing.T) {
	reloadCategories()
}

func TestMoveToRight(t *testing.T) {
	// case 1
	initData()
	MoveTo(gormMock, dresses, jackets, MoveDirectionRight)
	reloadCategories()

	assertCategoryEqual(t, clothing, 1, 22, 0, 0)
	assertCategoryEqual(t, mens, 2, 15, 1, clothing.ID)
	assertCategoryEqual(t, suits, 3, 14, 2, mens.ID)
	assertCategoryEqual(t, slacks, 4, 5, 3, suits.ID)
	assertCategoryEqual(t, jackets, 6, 7, 3, suits.ID)
	assertCategoryEqual(t, dresses, 8, 13, 3, suits.ID)
	assertCategoryEqual(t, eveningGowns, 9, 10, 4, dresses.ID)
	assertCategoryEqual(t, sunDresses, 11, 12, 4, dresses.ID)
	assertCategoryEqual(t, womens, 16, 21, 1, clothing.ID)
	assertCategoryEqual(t, skirts, 17, 18, 2, womens.ID)
	assertCategoryEqual(t, blouses, 19, 20, 2, womens.ID)

	// case 2
	initData()
	MoveTo(gormMock, suits, blouses, MoveDirectionRight)
	reloadCategories()

	assertCategoryEqual(t, clothing, 1, 22, 0, 0)
	assertCategoryEqual(t, mens, 2, 3, 1, clothing.ID)
	assertCategoryEqual(t, womens, 4, 21, 1, clothing.ID)
	assertCategoryEqual(t, dresses, 5, 10, 2, womens.ID)
	assertCategoryEqual(t, eveningGowns, 6, 7, 3, dresses.ID)
	assertCategoryEqual(t, sunDresses, 8, 9, 3, dresses.ID)
	assertCategoryEqual(t, skirts, 11, 12, 2, womens.ID)
	assertCategoryEqual(t, blouses, 13, 14, 2, womens.ID)
	assertCategoryEqual(t, suits, 15, 20, 2, womens.ID)
	assertCategoryEqual(t, slacks, 16, 17, 3, suits.ID)
	assertCategoryEqual(t, jackets, 18, 19, 3, suits.ID)
}

func TestMoveToLeft(t *testing.T) {
	// case 1
	initData()
	MoveTo(gormMock, dresses, jackets, MoveDirectionLeft)
	reloadCategories()

	assertCategoryEqual(t, clothing, 1, 22, 0, 0)
	assertCategoryEqual(t, mens, 2, 15, 1, clothing.ID)
	assertCategoryEqual(t, suits, 3, 14, 2, mens.ID)
	assertCategoryEqual(t, slacks, 4, 5, 3, suits.ID)
	assertCategoryEqual(t, dresses, 6, 11, 3, suits.ID)
	assertCategoryEqual(t, eveningGowns, 7, 8, 4, dresses.ID)
	assertCategoryEqual(t, sunDresses, 9, 10, 4, dresses.ID)
	assertCategoryEqual(t, jackets, 12, 13, 3, suits.ID)
	assertCategoryEqual(t, womens, 16, 21, 1, clothing.ID)
	assertCategoryEqual(t, skirts, 17, 18, 2, womens.ID)
	assertCategoryEqual(t, blouses, 19, 20, 2, womens.ID)

	// case 2
	initData()
	MoveTo(gormMock, suits, blouses, MoveDirectionLeft)
	reloadCategories()

	assertCategoryEqual(t, clothing, 1, 22, 0, 0)
	assertCategoryEqual(t, mens, 2, 3, 1, clothing.ID)
	assertCategoryEqual(t, womens, 4, 21, 1, clothing.ID)
	assertCategoryEqual(t, dresses, 5, 10, 2, womens.ID)
	assertCategoryEqual(t, eveningGowns, 6, 7, 3, dresses.ID)
	assertCategoryEqual(t, sunDresses, 8, 9, 3, dresses.ID)
	assertCategoryEqual(t, skirts, 11, 12, 2, womens.ID)
	assertCategoryEqual(t, suits, 13, 18, 2, womens.ID)
	assertCategoryEqual(t, slacks, 14, 15, 3, suits.ID)
	assertCategoryEqual(t, jackets, 16, 17, 3, suits.ID)
	assertCategoryEqual(t, blouses, 19, 20, 2, womens.ID)
}

func TestMoveToInner(t *testing.T) {
	// case 1
	initData()
	MoveTo(gormMock, mens, blouses, MoveDirectionInner)
	reloadCategories()

	assertCategoryEqual(t, clothing, 1, 22, 0, 0)
	assertCategoryEqual(t, womens, 2, 21, 1, clothing.ID)
	assertCategoryEqual(t, dresses, 3, 8, 2, womens.ID)
	assertCategoryEqual(t, eveningGowns, 4, 5, 3, dresses.ID)
	assertCategoryEqual(t, sunDresses, 6, 7, 3, dresses.ID)
	assertCategoryEqual(t, skirts, 9, 10, 2, womens.ID)
	assertCategoryEqual(t, blouses, 11, 20, 2, womens.ID)
	assertCategoryEqual(t, mens, 12, 19, 3, blouses.ID)
	assertCategoryEqual(t, suits, 13, 18, 4, mens.ID)
	assertCategoryEqual(t, slacks, 14, 15, 5, suits.ID)
	assertCategoryEqual(t, jackets, 16, 17, 5, suits.ID)

	// case 2
	initData()
	MoveTo(gormMock, skirts, slacks, MoveDirectionInner)
	reloadCategories()

	assertCategoryEqual(t, clothing, 1, 22, 0, 0)
	assertCategoryEqual(t, mens, 2, 11, 1, clothing.ID)
	assertCategoryEqual(t, suits, 3, 10, 2, mens.ID)
	assertCategoryEqual(t, slacks, 4, 7, 3, suits.ID)
	assertCategoryEqual(t, skirts, 5, 6, 4, slacks.ID)
	assertCategoryEqual(t, jackets, 8, 9, 3, suits.ID)
	assertCategoryEqual(t, womens, 12, 21, 1, clothing.ID)
	assertCategoryEqual(t, dresses, 13, 18, 2, womens.ID)
	assertCategoryEqual(t, eveningGowns, 14, 15, 3, dresses.ID)
	assertCategoryEqual(t, sunDresses, 16, 17, 3, dresses.ID)
	assertCategoryEqual(t, blouses, 19, 20, 2, womens.ID)
}

func assertCategoryEqual(t *testing.T, target Category, left, right, depth int, parentId int64) {
	assert.Equal(t, target.Lft, left)
	assert.Equal(t, target.Rgt, right)
	assert.Equal(t, target.Depth, depth)
	assert.Equal(t, target.ParentId, parentId)
}
