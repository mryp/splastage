package controllers

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mryp/splastage/app/models"
	"github.com/revel/revel"
	"gopkg.in/gorp.v1"
)

var (
	DbMap *gorp.DbMap
)

func InitDB() {
	revel.INFO.Println("gorp.InitDB()")
	db, err := sql.Open("sqlite3", "./stage.db")
	if err != nil {
		panic(err.Error())
	}
	DbMap = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// ここで好きにテーブルを定義する
	DbMap.AddTableWithName(models.Stage{}, "stage").SetKeys(true, "Id")

	//TODO: テストのためとりあえず全消し
	DbMap.DropTables()
	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		revel.INFO.Println("Error CreateTablesIfNotExists", err)
	}
}

type GorpController struct {
	*revel.Controller
	Transaction *gorp.Transaction
}

func (c *GorpController) Begin() revel.Result {
	revel.INFO.Println("gorp.Begin()")
	txn, err := DbMap.Begin() // ここで開始したtransactionをCOMMITする
	if err != nil {
		panic(err)
	}
	c.Transaction = txn
	return nil
}

func (c *GorpController) Commit() revel.Result {
	revel.INFO.Println("gorp.Comit()")
	if c.Transaction == nil {
		return nil
	}
	err := c.Transaction.Commit() // SQLによる変更をDBに反映
	if err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Transaction = nil // 正常終了した場合はROLLBACK処理に入らないようにする
	return nil
}

func (c *GorpController) Rollback() revel.Result {
	revel.INFO.Println("gorp.Rollback")
	if c.Transaction == nil {
		return nil
	}
	err := c.Transaction.Rollback() // 問題があった場合変更前の状態に戻す
	if err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Transaction = nil
	return nil
}
