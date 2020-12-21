package nested_set

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/bluele/factory-go/factory"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func databaseURL() string {
	databaseURL := os.Getenv("DATABASE_URL")
	if len(databaseURL) == 0 {
		databaseURL = "postgres://localhost:5432/test?sslmode=disable"
	}
	return databaseURL
}

var (
	memoryDB, _ = sql.Open("postgres", databaseURL())
	gormMock    = newMock(memoryDB)

	ctx = context.TODO()
)
var clothing, mens, suits, slacks, jackets, womens, dresses, skirts, blouses, eveningGowns, sunDresses Node

var NodeFactory = factory.NewFactory(&Node{
	Title: "Clothing",
	Rgt:   1,
	Lft:   2,
	Depth: 0,
}).
	OnCreate(func(args factory.Args) error {
		return gormMock.Create(args.Instance()).Error
	})

func newMock(_db *sql.DB) *gorm.DB {
	dir, _ := os.Getwd()
	os.MkdirAll(filepath.Join(dir, "./log"), 0777)
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
	gormMock.Exec("DROP TABLE IF EXISTS " + Node{}.TableName())
	err := gormMock.AutoMigrate(
		&Node{},
	)
	if err != nil {
		panic(err)
	}
	buildTestData()

}

func buildTestData() {
	clothing = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Clothing",
		"ParentId":      sql.NullInt64{Valid: false},
		"Lft":           1,
		"Rgt":           22,
		"Depth":         0,
		"ChildrenCount": 2,
	}).(*Node)
	mens = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Men's",
		"ParentId":      sql.NullInt64{Valid: true, Int64: clothing.ID},
		"Lft":           2,
		"Rgt":           9,
		"Depth":         1,
		"ChildrenCount": 1,
	}).(*Node)
	suits = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Suits",
		"ParentId":      sql.NullInt64{Valid: true, Int64: mens.ID},
		"Lft":           3,
		"Rgt":           8,
		"Depth":         2,
		"ChildrenCount": 2,
	}).(*Node)
	slacks = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Slacks",
		"ParentId": sql.NullInt64{Valid: true, Int64: suits.ID},
		"Lft":      4,
		"Rgt":      5,
		"Depth":    3,
	}).(*Node)
	jackets = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Jackets",
		"ParentId": sql.NullInt64{Valid: true, Int64: suits.ID},
		"Lft":      6,
		"Rgt":      7,
		"Depth":    3,
	}).(*Node)
	womens = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Women's",
		"ParentId":      sql.NullInt64{Valid: true, Int64: clothing.ID},
		"Lft":           10,
		"Rgt":           21,
		"Depth":         1,
		"ChildrenCount": 3,
	}).(*Node)
	dresses = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":         "Dresses",
		"ParentId":      sql.NullInt64{Valid: true, Int64: womens.ID},
		"Lft":           11,
		"Rgt":           16,
		"Depth":         2,
		"ChildrenCount": 2,
	}).(*Node)
	eveningGowns = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Evening Gowns",
		"ParentId": sql.NullInt64{Valid: true, Int64: dresses.ID},
		"Lft":      12,
		"Rgt":      13,
		"Depth":    3,
	}).(*Node)
	sunDresses = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Sun Dresses",
		"ParentId": sql.NullInt64{Valid: true, Int64: dresses.ID},
		"Lft":      14,
		"Rgt":      15,
		"Depth":    3,
	}).(*Node)
	skirts = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Skirts",
		"ParentId": sql.NullInt64{Valid: true, Int64: womens.ID},
		"Lft":      17,
		"Rgt":      18,
		"Depth":    2,
	}).(*Node)
	blouses = *NodeFactory.MustCreateWithOption(map[string]interface{}{
		"Title":    "Blouses",
		"ParentId": sql.NullInt64{Valid: true, Int64: womens.ID},
		"Lft":      19,
		"Rgt":      20,
		"Depth":    2,
	}).(*Node)
}

func reloadCategories() {
	clothing, _ = findNode(gormMock, clothing.ID)
	mens, _ = findNode(gormMock, mens.ID)
	suits, _ = findNode(gormMock, suits.ID)
	slacks, _ = findNode(gormMock, slacks.ID)
	jackets, _ = findNode(gormMock, jackets.ID)
	womens, _ = findNode(gormMock, womens.ID)
	dresses, _ = findNode(gormMock, dresses.ID)
	skirts, _ = findNode(gormMock, skirts.ID)
	blouses, _ = findNode(gormMock, blouses.ID)
	eveningGowns, _ = findNode(gormMock, eveningGowns.ID)
	sunDresses, _ = findNode(gormMock, sunDresses.ID)
}
