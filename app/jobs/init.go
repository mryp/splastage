package jobs

import (
	"github.com/mryp/splastage/app/controllers"
	"github.com/revel/modules/jobs/app/jobs"
	"github.com/revel/revel"
)

func init() {
	revel.OnAppStart(func() {
		revel.INFO.Println("call job.init")
		jobs.Schedule("0 * * * * *", jobs.Func(updateStageInfo))
	})
}

func updateStageInfo() {
	revel.INFO.Println("updateStageInfo")
	controllers.UpdateStageInfo()
}
