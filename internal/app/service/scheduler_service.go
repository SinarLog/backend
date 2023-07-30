// SCHEDULER is moved to its own repo because we were deploying it
// in a horizontally scalable engine and we do not want to replicate
// our scheduler. Do not expect the scheduler to work in this backend
// service, because we do not initialize it.

package service

type ISchedulerService interface{}
