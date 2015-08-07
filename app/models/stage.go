package models

import (
	"time"

	"github.com/revel/revel"

	"gopkg.in/gorp.v1"
)

type Stage struct {
	Id        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Rule      string    `db:"rule" json:"rule"`
	MatchType string    `db:"matchtype" json:"matchtype"`
	StartTime time.Time `db:"starttime" json:"starttime"`
	EndTime   time.Time `db:"endtime" json:"endtime"`
}

func StageInsertIfNotExists(dbmap *gorp.DbMap, stage Stage) error {
	err := dbmap.Insert(&stage)
	if err != nil {
		revel.INFO.Println("Insert error ", err)
	}

	return err
}

func StageSelectAll(dbmap *gorp.DbMap) []Stage {
	stageList := []Stage{}
	rows, _ := dbmap.Select(Stage{}, "select * from stage")
	for _, row := range rows {
		stage := row.(*Stage)
		stageList = append(stageList, *stage)
	}

	return stageList
}
