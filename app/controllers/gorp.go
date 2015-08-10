package controllers

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mryp/splastage/app/models"
	"github.com/revel/revel"
	"gopkg.in/gorp.v1"
)

var (
	//DbMap DB操作用オブジェクト
	DbMap *gorp.DbMap
)

//InitDB DBの初期化（オープン/テーブル生成）
func InitDB() {
	revel.INFO.Println("gorp.InitDB()")
	dbDriver, _ := revel.Config.String("db.driver")
	dbSpec, _ := revel.Config.String("db.spec")
	db, err := sql.Open(dbDriver, dbSpec)
	if err != nil {
		revel.ERROR.Panic(err.Error())
	}
	DbMap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	DbMap.AddTableWithName(models.Stage{}, "stage").SetKeys(true, "Id")
	DbMap.AddTableWithName(models.AccessLog{}, "accesslog").SetKeys(true, "Id")

	//DbMap.DropTables()
	err = DbMap.CreateTablesIfNotExists()
	if err != nil {
		revel.ERROR.Panic(err.Error())
	}
}

//GorpController 操作用コントローラー（各コントローラはこの構造体を組み込む）
type GorpController struct {
	*revel.Controller
	Transaction *gorp.Transaction
}

//Begin トランザクション開始（コントローラー処理時に内部で呼ぶため明示的に呼び出す必要なし）
func (c *GorpController) Begin() revel.Result {
	//revel.INFO.Println("gorp.Begin()")
	txn, err := DbMap.Begin() // ここで開始したtransactionをCOMMITする
	if err != nil {
		revel.ERROR.Panic(err.Error())
	}
	c.Transaction = txn
	return nil
}

//Commit トランザクションをコミット（コントローラー処理時に内部で呼ぶため明示的に呼び出す必要なし）
func (c *GorpController) Commit() revel.Result {
	//revel.INFO.Println("gorp.Comit()")
	if c.Transaction == nil {
		return nil
	}
	err := c.Transaction.Commit() // SQLによる変更をDBに反映
	if err != nil && err != sql.ErrTxDone {
		revel.ERROR.Panic(err.Error())
	}
	c.Transaction = nil // 正常終了した場合はROLLBACK処理に入らないようにする
	return nil
}

//Rollback トランザクション処理のロールバック（コントローラー処理時に内部で呼ぶため明示的に呼び出す必要なし）
func (c *GorpController) Rollback() revel.Result {
	//revel.INFO.Println("gorp.Rollback")
	if c.Transaction == nil {
		return nil
	}
	err := c.Transaction.Rollback() // 問題があった場合変更前の状態に戻す
	if err != nil && err != sql.ErrTxDone {
		revel.ERROR.Panic(err.Error())
	}
	c.Transaction = nil
	return nil
}
