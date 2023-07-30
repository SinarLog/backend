// SCHEDULER is moved to its own repo because we were deploying it
// in a horizontally scalable engine and we do not want to replicate
// our scheduler. Do not expect the scheduler to work in this backend
// service, because we do not initialize it.

package scheduler

import (
	"sync"

	"github.com/go-co-op/gocron"
	"sinarlog.com/internal/utils"
)

var (
	schedulerSingleInstance *gocron.Scheduler
	once                    sync.Once
)

func GetScheduler() *gocron.Scheduler {
	if schedulerSingleInstance == nil {
		once.Do(func() {
			schedulerSingleInstance = gocron.NewScheduler(utils.CURRENT_LOC).WaitForSchedule()
			schedulerSingleInstance.TagsUnique()
			schedulerSingleInstance.SingletonModeAll()
		})
	}

	return schedulerSingleInstance
}
