// SCHEDULER is moved to its own repo because we were deploying it
// in a horizontally scalable engine and we do not want to replicate
// our scheduler. Do not expect the scheduler to work in this backend
// service, because we do not initialize it.

package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"sinarlog.com/internal/delivery/v2/dto/mapper"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/utils"
)

var (
	once                    sync.Once
	schedulerSingleInstance *schedulerService
)

type schedulerService struct {
	db        *gorm.DB
	rdis      *redis.Client
	scheduler *gocron.Scheduler
}

func NewSchedulerService(scheduler *gocron.Scheduler, db *gorm.DB, rdis *redis.Client) *schedulerService {
	if schedulerSingleInstance == nil {
		once.Do(func() {
			schedulerSingleInstance = &schedulerService{db: db, rdis: rdis, scheduler: scheduler}

			if err := schedulerSingleInstance.RegisterInitJobs(); err != nil {
				log.Fatalf("Error while registering init jobs: %s", err.Error())
			}

			schedulerSingleInstance.scheduler.StartAsync()
		})
	}

	return schedulerSingleInstance
}

func (s *schedulerService) RegisterInitJobs() error {
	var errs error
	// Register init jobs
	jobs := []func() error{
		s.AttendanceCloserJob,
		s.LeaveRequestCloserJob,
		s.UpdateConfigNextDayJob,
		s.UpdateConfigNextMonthJob,
		s.OvertimeJobCloser,
	}

	for _, f := range jobs {
		if err := f(); err != nil {
			utils.AddError(errs, err)
		}
	}

	return errs
}

// AttendanceCloserJob closes an active attendances made by
// employees which runs every day and executes at 9:00 PM.
// On a closed attendance, it does not create overtime.
func (s *schedulerService) AttendanceCloserJob() error {
	task := func(job gocron.Job) {
		var errs error
		now := time.Now().In(utils.CURRENT_LOC)

		truee := true
		tx := s.db.WithContext(job.Context()).Begin()

		if err := tx.Exec(`UPDATE "attendances" SET clock_out_at = ?, closed_automatically = ?, updated_at = ?, done_for_the_day = ? WHERE done_for_the_day = ? OR clock_out_at = ?`,
			now,
			&truee,
			now,
			true,
			false,
			time.Time{},
		).Error; err != nil {
			tx.Rollback()
			log.Printf("Error while executing job %s\n", job.Tags()[0])
			utils.AddError(errs, err)
		}

		if err := tx.Exec(`UPDATE "employees" SET status = ?`, entity.UNAVAILABLE).Error; err != nil {
			tx.Rollback()
			log.Printf("Error while executing job %s while updating employees status\n", job.Tags()[0])
			errs = utils.AddError(errs, err)
		}

		if errs == nil {
			tx.Commit()
			log.Printf("Successfully ran job %s\n", job.Tags()[0])
		}
	}

	_, err := s.scheduler.Every(1).Day().At("21:00").Tag("attendanceCloser").WaitForSchedule().DoWithJobDetails(task)

	return err
}

// LeaveRequestCloserJob closes leave request where the requested
// leave starting date is 3 days after the current time the job
// executes. Any closed leave requests will be marked as closed
// and is not approved by any means. It also generates an automatic
// rejection reason.
func (s *schedulerService) LeaveRequestCloserJob() error {
	task := func(job gocron.Job) {
		var leaves []entity.Leave
		var rejectionReason string = "This leave request was closed because it had not finished processing 3 days before the start of request's date."

		// Collect all leaves
		if err := s.db.WithContext(job.Context()).Model(&entity.Leave{}).
			Where(`EXTRACT('day' FROM DATE_TRUNC('day', "leaves"."from" - now())) <= 3`).
			Where(`"leaves"."type" <> ?`, "SICK").
			Where("parent_id IS NULL").
			Where("closed_automatically IS NULL").
			Where(`
			(
				"leaves"."approved_by_manager" IS NULL
				AND
				"leaves"."approved_by_hr" IS NULL
			)
			OR
			(
				"leaves"."approved_by_manager" IS TRUE
				AND
				"leaves"."approved_by_manager" IS NULL
			)
			`).
			Preload("Childs").
			Find(&leaves).Error; err != nil {
			log.Printf("Unable to get leaves while executing job %s\n", job.Tags()[0])
			return
		}

		if len(leaves) == 0 {
			log.Printf("No leave requests to close\n")
		} else {
			// Iterate through the leaves
			for _, v := range leaves {
				var errs error
				// Close parent
				tx := s.db.WithContext(job.Context()).Begin()
				if err := tx.Exec(`UPDATE "leaves" SET closed_automatically = ?, rejection_reason = ? WHERE id = ?`, true, rejectionReason, v.Id).Error; err != nil {
					tx.Rollback()
					errs = utils.AddError(errs, err)
					log.Printf("Error while closing leave of id %s on *schedulerService.(LeaveRequestCloserJob): %s\n", v.Id, err)
				}
				sql := generateUpdateLeaveQuotaSql(v.Type, true)
				if err := tx.Exec(sql, utils.CountNumberOfWorkingDays(v.From, v.To), v.EmployeeID).Error; err != nil {
					tx.Rollback()
					errs = utils.AddError(errs, err)
					log.Printf("Error while returning quota of employee_id %s with leave_id %s on *schedulerService.(LeaveRequestCloserJob): %s\n", v.EmployeeID, v.Id, err)
				}

				// Close childs only if pending
				for _, c := range v.Childs {
					if mapper.LeaveStatusMapper(c) == "PENDING" {
						if err := tx.Exec(`UPDATE "leaves" SET closed_automatically = ?, rejection_reason = ? WHERE id = ?`, true, rejectionReason, c.Id).Error; err != nil {
							tx.Rollback()
							errs = utils.AddError(errs, err)
							log.Printf("Error while closing leave of id %s on *schedulerService.(LeaveRequestCloserJob): %s\n", c.Id, err)
						}

						sql := generateUpdateLeaveQuotaSql(c.Type, true)
						if err := tx.Exec(sql, utils.CountNumberOfWorkingDays(c.From, c.To), v.EmployeeID).Error; err != nil {
							tx.Rollback()
							errs = utils.AddError(errs, err)
							log.Printf("Error while returning quota of employee_id %s with leave_id %s on *schedulerService.(LeaveRequestCloserJob): %s\n", v.EmployeeID, c.Id, err)
						}

						log.Printf("Successfully closed leave_id %s\n", c.Id)
					}
				}

				if errs == nil {
					tx.Commit()
					log.Printf("Successfully closed leave_id %s\n", v.Id)
				}
			}
		}

		log.Printf("Finished running job %s\n", job.Tags()[0])
	}

	_, err := s.scheduler.Every(1).Day().At("00:00").Tag("leaveCloser").WaitForSchedule().DoWithJobDetails(task)

	return err
}

func (s *schedulerService) UpdateConfigNextDayJob() error {
	task := func(job gocron.Job) {
		// Checks if the key exists
		exists, err := s.rdis.Exists(job.Context(), entity.CONFIG_KEY_NEXT_DAY).Result()
		if err != nil {
			log.Printf("Unable to check if key exists in redis for *schedulerService.(UpdateConfigNextDayJob): %s\n", err)
			return
		}

		if exists == 1 {
			var configToChange entity.Configuration
			var config entity.Configuration

			// Get config to change from redis
			if err := s.rdis.HGetAll(job.Context(), entity.CONFIG_KEY_NEXT_DAY).Scan(&configToChange); err != nil {
				log.Printf("Unable to retrieve config hash from redis: %s\n", err)
				return
			}

			// Query current config
			tx := s.db.WithContext(job.Context()).Begin()
			if err := tx.Model(&config).First(&config).Error; err != nil {
				log.Printf("Unable to get config: %s\n", err)
				tx.Rollback()
				return
			}

			// Apply changes
			config.AcceptanceAttendanceInterval = configToChange.AcceptanceAttendanceInterval
			config.OfficeStartTime = time.Date(
				config.OfficeStartTime.Year(),
				config.OfficeStartTime.Month(),
				config.OfficeStartTime.Day(),
				configToChange.OfficeStartTimeHour,
				configToChange.OfficeStartTimeMinute,
				0,
				0,
				utils.CURRENT_LOC,
			)
			config.OfficeEndTime = time.Date(
				config.OfficeEndTime.Year(),
				config.OfficeEndTime.Month(),
				config.OfficeEndTime.Day(),
				configToChange.OfficeEndTimeHour,
				configToChange.OfficeEndTimeMinute,
				0,
				0,
				utils.CURRENT_LOC,
			)

			// Save changes
			if err := tx.Model(&config).Save(&config).Error; err != nil {
				log.Printf("Unable to save config changes for *schedulerService.(UpdateConfigNextDayJob): %s\n", err)
				tx.Rollback()
				return
			}

			if s.rdis.Del(job.Context(), entity.CONFIG_KEY_NEXT_DAY).Val() == 0 {
				log.Printf("Unable to delete redis key at *schedulerService.(UpdateConfigNextDayJob) for %s\n", entity.CONFIG_KEY_NEXT_DAY)
			}
			tx.Commit()
			log.Printf("Successfully ran job %s\n", job.Tags()[0])
			return
		} else {
			log.Printf("Nothing to update for *schedulerService.(UpdateConfigNextDayJob)\n")
			return
		}
	}

	_, err := s.scheduler.Every(1).Day().At("00:00").Tag("configNextDayUpdater").WaitForSchedule().DoWithJobDetails(task)

	return err
}

func (s *schedulerService) UpdateConfigNextMonthJob() error {
	task := func(job gocron.Job) {
		// Checks if the key exists
		exists, err := s.rdis.Exists(job.Context(), entity.CONFIG_KEY_NEXT_MONTH).Result()
		if err != nil {
			log.Printf("Unable to check if key exists in redis for *schedulerService.(UpdateConfigNextMonthJob): %s\n", err)
			return
		}

		if exists == 1 {
			var configToChange entity.Configuration
			var config entity.Configuration

			// Get config to change from redis
			if err := s.rdis.HGetAll(job.Context(), entity.CONFIG_KEY_NEXT_MONTH).Scan(&configToChange); err != nil {
				log.Printf("Unable to retrieve config hash from redis: %s\n", err)
				return
			}

			// Query current config
			tx := s.db.WithContext(job.Context()).Begin()
			if err := tx.Model(&config).First(&config).Error; err != nil {
				log.Printf("Unable to get config: %s\n", err)
				tx.Rollback()
				return
			}

			// Apply changes
			config.AcceptanceLeaveInterval = configToChange.AcceptanceLeaveInterval
			config.DefaultYearlyQuota = configToChange.DefaultYearlyQuota
			config.DefaultMarriageQuota = configToChange.DefaultMarriageQuota

			// Save changes
			if err := tx.Model(&config).Save(&config).Error; err != nil {
				log.Printf("Unable to save config changes for *schedulerService.(UpdateConfigNextMonthJob): %s\n", err)
				tx.Rollback()
				return
			}

			if s.rdis.Del(job.Context(), entity.CONFIG_KEY_NEXT_MONTH).Val() == 0 {
				log.Printf("Unable to delete redis key at *schedulerService.(UpdateConfigNextMonthJob) for %s\n", entity.CONFIG_KEY_NEXT_MONTH)
			}
			tx.Commit()
			log.Printf("Successfully ran job %s\n", job.Tags()[0])
			return
		} else {
			log.Printf("Nothing to update for *schedulerService.(UpdateConfigNextMonthJob)\n")
			return
		}
	}

	_, err := s.scheduler.Cron("0 0 1 * *").Tag("configNextMonthUpdater").WaitForSchedule().DoWithJobDetails(task)

	return err
}

func (s *schedulerService) OvertimeJobCloser() error {
	task := func(job gocron.Job) {
		var overtimes []entity.Overtime
		truee := true

		// Query all the overtimes that has not been processed
		if err := s.db.WithContext(job.Context()).
			Model(&entity.Overtime{}).
			Where("approved_by_manager IS NULL AND action_by_manager_at IS NULL").
			Preload("Attendance.Employee").
			Find(&overtimes).Error; err != nil {
			log.Printf("Unable to query all pending overtimes for *schedulerService.(OvertimeJobCloser): %s\n", err)
			return
		}

		if len(overtimes) == 0 {
			log.Println("No overtimes submissions to close")
		} else {
			// Close each overtimes
			for _, v := range overtimes {
				tx := s.db.WithContext(job.Context()).Begin()
				v.ClosedAutomatically = &truee
				v.RejectionReason = "This overtime submission is closed because it was not processed until 24th of the month."
				if err := tx.Model(&v).Save(&v).Error; err != nil {
					tx.Rollback()
					continue
				}
				tx.Commit()
			}

			log.Printf("Finished running job %s\n", job.Tags()[0])
		}
	}

	_, err := s.scheduler.Cron("0 0 24 * *").Tag("overtimeCloser").WaitForSchedule().DoWithJobDetails(task)

	return err
}

func generateUpdateLeaveQuotaSql(leaveType entity.LeaveType, reverse bool) string {
	mapper := map[entity.LeaveType]string{
		entity.ANNUAL:   "yearly_count",
		entity.MARRIAGE: "marriage_count",
		entity.UNPAID:   "unpaid_count",
	}

	switch leaveType {
	case entity.UNPAID:
		if reverse {
			return fmt.Sprintf("UPDATE employee_leaves_quota SET %s = %s - ? WHERE employee_id = ?", mapper[leaveType], mapper[leaveType])
		}
		return fmt.Sprintf("UPDATE employee_leaves_quota SET %s = %s + ? WHERE employee_id = ?", mapper[leaveType], mapper[leaveType])
	case entity.ANNUAL, entity.MARRIAGE:
		if reverse {
			return fmt.Sprintf("UPDATE employee_leaves_quota SET %s = %s + ? WHERE employee_id = ?", mapper[leaveType], mapper[leaveType])
		}
		return fmt.Sprintf("UPDATE employee_leaves_quota SET %s = %s - ? WHERE employee_id = ?", mapper[leaveType], mapper[leaveType])
	default:
		return ""
	}
}
