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
//bool = 追加が行われたかどうか
//error = 追加処理時に発生したエラー（エラーがないときはnil）
func StageInsertIfNotExists(dbmap *gorp.DbMap, stage Stage) (bool, error) {
	if StageIsExists(dbmap, stage) {
		return false, nil
	}
	err := dbmap.Insert(&stage)
	if err != nil {
		revel.INFO.Println("Insert error ", err)
		return false, err
	} else {
		return true, nil
	}
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

//最後のステージ情報を取得する
func StageSelectLast(dbmap *gorp.DbMap) []Stage {
	var lastStage Stage
	err := dbmap.SelectOne(&lastStage, "select * from stage order by endtime desc limit 1")
	if err != nil {
		revel.INFO.Println("StageSelectLast SelectOne", err)
		return nil
	}

	var stageList []Stage
	_, err2 := dbmap.Select(&stageList, "select * from stage where endtime=?", lastStage.EndTime)
	if err2 != nil {
		revel.INFO.Println("Select error", err)
	}

	return stageList
}
