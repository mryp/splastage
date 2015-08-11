package models

import (
	"time"

	"github.com/revel/revel"

	"gopkg.in/gorp.v1"
)

//アクセスログ情報
type AccessLog struct {
	Id        int64     `db:"id"`
	Action    string    `db:"action"`
	ClientID  string    `db:"clientid"`
	Host      string    `db:"host"`
	UserAgent string    `db:"useragent"`
	Uptime    time.Time `db:"uptime"`
}

func AccessLogCreate(action string, clientId string, request *revel.Request) AccessLog {
	//timeをDBに保存すると必ずUTCで保存されるようなのでUTCをJSTと一致させる
	now := time.Now().UTC().Add(9 * time.Hour)
	revel.INFO.Println("now time", now, now.UTC())
	return AccessLog{Action: action, ClientID: clientId, Host: request.RemoteAddr, UserAgent: request.UserAgent(), Uptime: now}
}

func AccessLogInsert(dbmap *gorp.DbMap, accessLog AccessLog) error {
	err := dbmap.Insert(&accessLog)
	if err != nil {
		revel.WARN.Println("データ追加失敗", err)
		return err
	}

	return nil
}

func AccessLogSelect(dbmap *gorp.DbMap, startTime time.Time, endTime time.Time) []AccessLog {
	startText := startTime.Format("2006-01-02 15:04:05")
	endText := endTime.Format("2006-01-02 15:04:05")
	var logList []AccessLog
	_, err := dbmap.Select(&logList, "select * from accesslog where uptime between ? and ?", startText, endText)
	if err != nil {
		revel.WARN.Println("データ取得失敗", err)
		return nil
	}
	return logList
}

func AccessLogSelectToday(dbmap *gorp.DbMap) []AccessLog {
	now := time.Now()
	startText := now.Format("2006-01-02")
	endText := now.Add(24 * time.Hour).Format("2006-01-02")
	var logList []AccessLog
	_, err := dbmap.Select(&logList, "select * from accesslog where uptime between ? and ?", startText, endText)
	if err != nil {
		revel.WARN.Println("データ取得失敗", err)
		return nil
	}
	return logList
}
