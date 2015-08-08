package models

import (
	"time"

	"github.com/revel/revel"

	"gopkg.in/gorp.v1"
)

//ステージ情報
type Stage struct {
	Id        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Rule      string    `db:"rule" json:"rule"`
	MatchType string    `db:"matchtype" json:"matchtype"`
	StartTime time.Time `db:"starttime" json:"starttime"`
	EndTime   time.Time `db:"endtime" json:"endtime"`
}

//ステージ情報を追加する
//すでに存在するときは追加しない
func StageInsertIfNotExists(dbmap *gorp.DbMap, stage Stage) error {
	if StageIsExists(dbmap, stage) {
		return nil
	}
	err := dbmap.Insert(&stage)
	if err != nil {
		revel.INFO.Println("Insert error ", err)
	}

	return err
}

//指定したステージ情報がすでに保存されているかどうか
func StageIsExists(dpmap *gorp.DbMap, stage Stage) bool {
	var output Stage
	err := dpmap.SelectOne(&output, "select * from stage where name=? and rule=? and matchtype=? and starttime=? and endtime=?",
		stage.Name, stage.Rule, stage.MatchType, stage.StartTime, stage.EndTime)
	if err != nil {
		//取得できなかったのでデータなし
		return false
	}

	//データ発見
	return true
}

//現在保存されているステージ情報をすべて取得する
func StageSelectAll(dbmap *gorp.DbMap) []Stage {
	var stageList []Stage
	_, err := dbmap.Select(&stageList, "select * from stage")
	if err != nil {
		revel.INFO.Println("Select error ", err)
	}
	return stageList
}
