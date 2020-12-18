package nested_set

import (
	"context"
	"database/sql"
	"github.com/bluele/factory-go/factory"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path/filepath"
)

func databaseURL() string {
	return "postgres://localhost:5432/go_nested_set_test?sslmode=disable"
}

var (
	memoryDB, _ = sql.Open("postgres", databaseURL())
	gormMock    = newMock(memoryDB)

	ctx = context.TODO()
)
var clothing, mens, suits, slacks, jackets, womens, dresses, skirts, blouses, eveningGowns, sunDresses Category

var CategoryFactory = factory.NewFactory(&Category{
	Title:    "Clothing",
	ParentId: 0,
	Rgt:      1,
	Lft:      2,
	Depth:    0,
}).
	OnCreate(func(args factory.Args) error {
		return gormMock.Create(args.Instance()).Error
	})

func newMock(_db *sql.DB) *gorm.DB {
	dir, _ := os.Getwd()
	logFile, err := os.OpenFile(filepath.Join(dir, "./log/test.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		DSN:                  databaseURL(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.New(log.New(logFile, "\n", 1), logger.Config{
			LogLevel: logger.Info,
		}),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	return gormDB
}

func initData() {
	gormMock.Exec("DROP TABLE IF EXISTS " + Category{}.TableName())
	err := gormMock.AutoMigrate(
		&Category{},
	)
	if err != nil {
		panic(err)
	}
	buildTestData()

}

func buildTestData() {
	clothing = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title": "Clothing",
		"Lft":   1,
		"Rgt":   22,
		"Depth": 0,
	}).(*Category)
	mens = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Men's",
		"ParentId": clothing.ID,
		"Lft":      2,
		"Rgt":      9,
		"Depth":    1,
	}).(*Category)
	suits = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Suits",
		"ParentId": mens.ID,
		"Lft":      3,
		"Rgt":      8,
		"Depth":    2,
	}).(*Category)
	slacks = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Slacks",
		"ParentId": suits.ID,
		"Lft":      4,
		"Rgt":      5,
		"Depth":    3,
	}).(*Category)
	jackets = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Jackets",
		"ParentId": suits.ID,
		"Lft":      6,
		"Rgt":      7,
		"Depth":    3,
	}).(*Category)
	womens = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Women's",
		"ParentId": clothing.ID,
		"Lft":      10,
		"Rgt":      21,
		"Depth":    1,
	}).(*Category)
	dresses = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Dresses",
		"ParentId": womens.ID,
		"Lft":      11,
		"Rgt":      16,
		"Depth":    2,
	}).(*Category)
	eveningGowns = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Evening Gowns",
		"ParentId": dresses.ID,
		"Lft":      12,
		"Rgt":      13,
		"Depth":    3,
	}).(*Category)
	sunDresses = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Sun Dresses",
		"ParentId": dresses.ID,
		"Lft":      14,
		"Rgt":      15,
		"Depth":    3,
	}).(*Category)
	skirts = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Skirts",
		"ParentId": womens.ID,
		"Lft":      17,
		"Rgt":      18,
		"Depth":    2,
	}).(*Category)
	blouses = *CategoryFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Blouses",
		"ParentId": womens.ID,
		"Lft":      19,
		"Rgt":      20,
		"Depth":    2,
	}).(*Category)
}

func reloadCategories() {
	clothing, _ = findCategory(gormMock, clothing.ID)
	mens, _ = findCategory(gormMock, mens.ID)
	suits, _ = findCategory(gormMock, suits.ID)
	slacks, _ = findCategory(gormMock, slacks.ID)
	jackets, _ = findCategory(gormMock, jackets.ID)
	womens, _ = findCategory(gormMock, womens.ID)
	dresses, _ = findCategory(gormMock, dresses.ID)
	skirts, _ = findCategory(gormMock, skirts.ID)
	blouses, _ = findCategory(gormMock, blouses.ID)
	eveningGowns, _ = findCategory(gormMock, eveningGowns.ID)
	sunDresses, _ = findCategory(gormMock, sunDresses.ID)
}
