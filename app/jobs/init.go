package jobs

import (
	"strconv"

	"github.com/mryp/splastage/app/controllers"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
)

func init() {
	revel.OnAppStart(func() {
		revel.INFO.Println("call job.init")

		//設定方法：https://github.com/revel/cron/blob/master/doc.go
		jobs.Now(jobs.Func(updateStageInfoInit))
		jobs.Schedule("0 1 * * * *", jobs.Func(updateStageInfoFromSchedule))
	})
}

//初回起動時の最新データ取得処理
func updateStageInfoInit() {
	revel.INFO.Println("起動時ステージ情報取得開始")
	ret := controllers.UpdateStageFromIkaring()
	revel.INFO.Println("更新結果 ret=" + strconv.FormatBool(ret))
}

//指定時間ごとの定期的データ取得処理
func updateStageInfoFromSchedule() {
	revel.INFO.Println("スケジュールステージ情報取得開始")
	ret := controllers.UpdateStageFromIkaring()
	revel.INFO.Println("更新結果 ret=" + strconv.FormatBool(ret))
}
