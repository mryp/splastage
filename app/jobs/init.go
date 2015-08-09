package jobs

import (
	"github.com/mryp/splastage/app/controllers"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
)

func init() {
	revel.OnAppStart(func() {
		revel.INFO.Println("call job.init")

		//設定方法：https://github.com/revel/cron/blob/master/doc.go
		jobs.Now(jobs.Func(updateStageInfo))
		jobs.Schedule("0 1 * * * *", jobs.Func(updateStageInfo))
	})
}

func updateStageInfo() {
	revel.INFO.Println("updateStageInfo")
	controllers.UpdateStageInfo()
}
