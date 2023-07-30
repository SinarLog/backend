package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/app/service"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type attendanceUseCase struct {
	attRepo      repo.IAttendanceRepo
	leaveRepo    repo.ILeaveRepo
	configRepo   repo.IConfigRepo
	emplRepo     repo.IEmployeeRepo
	dkService    service.IDoorkeeperService
	mailService  service.IMailerService
	notifService service.INotifService
}

func NewAttendaceUseCase(
	attRepo repo.IAttendanceRepo,
	leaveRepo repo.ILeaveRepo,
	configRepo repo.IConfigRepo,
	emplRepo repo.IEmployeeRepo,
	dkService service.IDoorkeeperService,
	mailService service.IMailerService,
	notifService service.INotifService,
) *attendanceUseCase {
	return &attendanceUseCase{
		attRepo:      attRepo,
		leaveRepo:    leaveRepo,
		configRepo:   configRepo,
		emplRepo:     emplRepo,
		dkService:    dkService,
		mailService:  mailService,
		notifService: notifService,
	}
}

/*
*********************************
ACTOR: STAFF and MANAGER
*********************************
*/

func (uc *attendanceUseCase) RequestClockIn(ctx context.Context, employee entity.Employee) error {
	// Checks if the employee has clocked in today
	hasClockedIn, err := uc.attRepo.EmployeeHasClockedInToday(ctx, employee.Id)
	if err == nil {
		if hasClockedIn {
			return NewDomainError("Attendance", fmt.Errorf("employee has clocked in for today"))
		}
	} else {
		return NewRepositoryError("Attendance", err)
	}

	// Checks if the employee is on leave
	onLeave, err := uc.leaveRepo.EmployeeIsOnLeaveToday(ctx, employee.Id)
	if err != nil {
		return NewRepositoryError("Leave", utils.AddError(fmt.Errorf("unable to identify if the employee today is on leave"), err))
	}
	if employee.Status == entity.ON_LEAVE || onLeave {
		return NewDomainError("Attendance", fmt.Errorf("unable to clock in if employee is on leave"))
	}

	// Query the office configuration
	config, err := uc.configRepo.GetConfiguration(ctx)
	if err != nil {
		return NewRepositoryError("Configuration", err)
	}

	// Checks if clock in request is made after office end time
	now := time.Now().In(utils.CURRENT_LOC)
	dur, _ := time.ParseDuration(config.AcceptanceAttendanceInterval)
	officeEndTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		config.OfficeEndTime.Hour(),
		config.OfficeEndTime.Minute(),
		config.OfficeEndTime.Second(),
		config.OfficeEndTime.Nanosecond(),
		utils.CURRENT_LOC,
	)
	if now.After(officeEndTime.Add(-dur)) {
		return NewDomainError("Attendance", fmt.Errorf("clocking in after office's end time is not allowed"))
	}

	// Generate OTP
	otp, timestamp, exp := uc.dkService.GenerateOTP()
	if err := uc.attRepo.SaveClockInOTPTimestamp(ctx, employee.Id, timestamp, exp); err != nil {
		return NewRepositoryError("Attendance", fmt.Errorf("unable to save otp: %w", err))
	}

	// Send to email.
	data := map[string]any{
		"FullName": employee.FullName,
		"OTP":      otp,
		"Action":   "Clock In",
		"Exp":      fmt.Sprint(exp),
	}
	go uc.sendClockInOTPMail(employee.Email, data)

	return nil
}

func (uc *attendanceUseCase) ClockIn(ctx context.Context, employee entity.Employee, req vo.ClockInRequest) error {
	// Checks whether employee has clocked in today
	hasClockedIn, err := uc.attRepo.EmployeeHasClockedInToday(ctx, employee.Id)
	if err == nil {
		if hasClockedIn {
			return NewDomainError("Attendance", fmt.Errorf("employee has clocked in for today"))
		}
	} else {
		return NewRepositoryError("Attendance", err)
	}

	// Validate OTP
	timestamp, err := uc.attRepo.GetClockInOTPTimestamp(ctx, employee.Id)
	if err == nil {
		match := uc.dkService.VerifyOTP(req.OTP, timestamp)
		if !match {
			return NewClientError("Attendance", fmt.Errorf("OTP does not match"))
		}
	} else {
		return NewRepositoryError("Attendance", err)
	}

	// Query the office configuration
	config, err := uc.configRepo.GetConfiguration(ctx)
	if err != nil {
		return NewRepositoryError("Configuration", err)
	}

	// Create and validate attendance
	attendance := entity.Attendance{
		EmployeeID:    employee.Id,
		Employee:      employee,
		ClockInAt:     time.Now().In(utils.CURRENT_LOC),
		DoneForTheDay: false,
		ClockInLoc:    req.Loc,
	}
	if err := attendance.ValidateClockIn(config); err != nil {
		return NewDomainError("Attendance", err)
	}

	// Checks whether it is a late clock in
	if attendance.IsLateClockIn(config) {
		attendance.LateClockIn = true
	}

	// Persist record
	if err := uc.attRepo.CreateNewAttendance(ctx, attendance); err != nil {
		return NewRepositoryError("Attendance", err)
	}

	// Set employee to available
	if err := uc.emplRepo.SetEmployeeStatusTo(ctx, employee.Id, entity.AVAILABLE); err != nil {
		return NewErrorWithReport(
			"Employee",
			500,
			ErrUnexpected,
			fmt.Errorf("unable to set your status to available. But don't worry your attendance is successfully saved"),
			"Please report this issue to support@sinarlog.com",
		)
	}

	return nil
}

func (uc *attendanceUseCase) RetrieveTodaysAttendance(ctx context.Context, employee entity.Employee) (entity.Attendance, error) {
	attendance, err := uc.attRepo.GetTodaysAttendanceByEmployeeId(ctx, employee.Id)
	if err != nil {
		return entity.Attendance{}, NewRepositoryError("Attendance", err)
	}

	return attendance, nil
}

// RequestClockOut checks all requirements for clocking out. On successfull, it sends a
// response indicating whether the attendance is an overtime or not. It also indicates
// whether an overtime submission is possible or not given a limit in overtime duration.
// When a manager is requesting clockout, immediately return an empty report. This means,
// only staff are able to have overtime.
func (uc *attendanceUseCase) RequestClockOut(ctx context.Context, employee entity.Employee) (entity.OvertimeOnAttendanceReport, error) {
	// Checks if there are any active attendance
	hasActiveAttendance, err := uc.attRepo.EmployeeHasActiveAttendance(ctx, employee.Id)
	if err == nil {
		if !hasActiveAttendance {
			return entity.OvertimeOnAttendanceReport{}, NewClientError("Attendance", fmt.Errorf("you have no active attendance"))
		}
	} else {
		return entity.OvertimeOnAttendanceReport{}, NewRepositoryError("Attendance", err)
	}

	// Checks the actor...
	if employee.ManagerID == nil {
		return entity.OvertimeOnAttendanceReport{}, nil
	}

	// Query the attendance
	attendance, err := uc.attRepo.GetActiveAttendanceByEmployeeId(ctx, employee.Id)
	if err != nil {
		return entity.OvertimeOnAttendanceReport{}, NewRepositoryError("Attendance", err)
	}

	// Checks if the attendance is an overtime
	report, ucErr := uc.createOvertimeOnAttendanceReport(ctx, attendance)
	if err != nil {
		return entity.OvertimeOnAttendanceReport{}, ucErr
	}

	return report, nil
}

func (uc *attendanceUseCase) ClockOut(ctx context.Context, employee entity.Employee, payload vo.ClockOutPayload) error {
	// Checks if there are any active attendance
	hasActiveAttendance, err := uc.attRepo.EmployeeHasActiveAttendance(ctx, employee.Id)
	if err == nil {
		if !hasActiveAttendance {
			return NewClientError("Attendance", fmt.Errorf("you have no active attendance"))
		}
	} else {
		return NewRepositoryError("Attendance", err)
	}

	// Query the attendance
	attendance, err := uc.attRepo.GetActiveAttendanceByEmployeeId(ctx, employee.Id)
	if err != nil {
		return NewRepositoryError("Attendance", err)
	}
	attendance.Employee = employee

	// Query configurations
	config, err := uc.configRepo.GetConfiguration(ctx)
	if err != nil {
		return NewRepositoryError("Config", err)
	}

	// Modify and validate attendance
	attendance.ClockOutAt = time.Now().In(utils.CURRENT_LOC)
	attendance.ClockOutLoc = payload.Loc
	attendance.DoneForTheDay = true
	attendance.EarlyClockOut = attendance.IsEarlyClockOut(config)
	attendance.Employee.Status = entity.UNAVAILABLE
	if err := attendance.ValidateClockOut(); err != nil {
		return NewDomainError("Attendance", err)
	}

	report, err := uc.createOvertimeOnAttendanceReport(ctx, attendance)
	if err != nil {
		return NewRepositoryError("Attendance", err)
	}

	// If should create an overtime record and the user confirms and the actor is a staff...
	if report.ShouldCreateOvertimeRecord() && payload.Confirmation && employee.ManagerID != nil {
		attendance.Overtime = &entity.Overtime{
			AttendanceID: attendance.Id,
			Duration:     int(report.OvertimeAcceptedDuration),
			Reason:       payload.Reason,
			ManagerID:    attendance.Employee.ManagerID,
		}
		if err := attendance.Overtime.Validate(); err != nil {
			return NewDomainError("Overtime", err)
		}

		if err := uc.attRepo.CloseAttendance(ctx, attendance); err != nil {
			return NewRepositoryError("Attendance", err)
		}

		// Query the manager to be notifed about the overtime submission
		manager, err := uc.emplRepo.GetEmployeeById(ctx, *attendance.Employee.ManagerID)
		if err != nil {
			return NewErrorWithReport(
				"Notification",
				207,
				ErrUnexpected,
				fmt.Errorf("unable to query manager"),
				"We were not able to notify your manager about the overtime submission. However, your attendance and overtime submission has been successfully saved",
			)
		}

		// Sends and checks if anyone recieve the notification
		wasReceived, err := uc.notifService.SendOvertimeSubmissionNotification(ctx, manager, employee)
		if err != nil {
			log.Printf("unable to publish notification")
		} else if wasReceived == 0 {
			// If no one recieves, sends via email instead
			go uc.sendOvertimeSubmissionEmail(manager, attendance)
		}
	} else {
		if err := uc.attRepo.CloseAttendance(ctx, attendance); err != nil {
			return NewRepositoryError("Attendance", err)
		}
	}

	return nil
}

func (uc *attendanceUseCase) RetrieveMyAttendanceHistory(ctx context.Context, employee entity.Employee, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	attendances, page, err := uc.attRepo.GetMyAttendancesHistory(ctx, employee.Id, q)
	if err != nil {
		return nil, page, NewRepositoryError("Attendances", err)
	}

	return attendances, page, nil
}

func (uc *attendanceUseCase) createOvertimeOnAttendanceReport(ctx context.Context, attendance entity.Attendance) (entity.OvertimeOnAttendanceReport, error) {
	// Differentiate the case for weekend and weekday
	switch attendance.ClockInAt.In(utils.CURRENT_LOC).Weekday() {
	case time.Saturday, time.Sunday:
		// Any attendance made on weekend is an overtime.
		// Returns the report right away.
		report := entity.OvertimeOnAttendanceReport{
			IsOvertime:               true,
			IsOnHoliday:              true,
			OvertimeAcceptedDuration: time.Since(attendance.ClockInAt),
		}
		return report, nil
	default:
		// Query configurations
		config, err := uc.configRepo.GetConfiguration(ctx)
		if err != nil {
			return entity.OvertimeOnAttendanceReport{}, NewRepositoryError("Config", err)
		}

		// Set the time breakpoint
		now := time.Now().In(utils.CURRENT_LOC)
		workDur := now.Sub(attendance.ClockInAt)
		officeWorkDur := config.OfficeWorkDuration()
		var report entity.OvertimeOnAttendanceReport

		// Indicates if the attendance duration is more than office work duration
		if workDur > officeWorkDur {
			report.IsOvertime = true
			report.IsOvertimeAvailable = true
			report.OvertimeDuration = workDur - officeWorkDur
			report.MaxAllowedDailyDuration = time.Duration(config.MaxOvertimeDailyDur) * time.Hour
			report.MaxAllowedWeeklyDuration = time.Duration(config.MaxOvertimeWeeklyDur) * time.Hour

			// Checks if the attendance is more than the allowed daily duration
			if report.OvertimeDuration > report.MaxAllowedDailyDuration {
				report.OvertimeAcceptedDuration = report.MaxAllowedDailyDuration
				report.IsOvertimeLeakage = true
			} else {
				report.OvertimeAcceptedDuration = report.OvertimeDuration
			}

			// Query the employee overtime for this week
			sum, err := uc.attRepo.SumWeeklyOvertimeDurationByEmployeeId(ctx, attendance.EmployeeID)
			if err != nil {
				return report, NewRepositoryError("Attendance", err)
			}
			weeklySum := time.Duration(sum)

			// Checks whether the weekly sum is more than max allowed weekly duration
			if weeklySum >= report.MaxAllowedWeeklyDuration {
				// Case where weekly sum is equal or more than max weekly dur
				report.IsOvertimeAvailable = false
				report.IsOvertimeLeakage = true
			} else if report.OvertimeAcceptedDuration > report.MaxAllowedWeeklyDuration-weeklySum {
				// Case where the daily dur is more than the remaining weekly dur
				report.IsOvertimeLeakage = true
				report.IsOvertimeAvailable = true
				report.OvertimeAcceptedDuration = report.MaxAllowedWeeklyDuration - weeklySum
			}
		}
		return report, nil
	}
}

/*
*********************************
ACTOR: MANAGER
*********************************
*/

func (uc *attendanceUseCase) RetrieveMyStaffsAttendanceHistory(ctx context.Context, manager entity.Employee, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	if _, err := q.CommonQuery.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Attendance", err)
	}

	q.CommonQuery.Pagination.Order = "clock_in_at"
	attendances, page, err := uc.attRepo.GetStaffsAttendancesHistory(ctx, manager.Id, q)
	if err != nil {
		return nil, page, NewRepositoryError("Attendance", err)
	}

	return attendances, page, nil
}

func (uc *attendanceUseCase) RetrieveStaffAttendanceHistory(ctx context.Context, manager entity.Employee, employeeId string, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	if _, err := q.CommonQuery.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Attendance", err)
	}

	// Verify that the employee is indeed under the manager
	employee, err := uc.emplRepo.GetEmployeeById(ctx, employeeId)
	if err != nil {
		return nil, vo.PaginationDTOResponse{}, NewRepositoryError("Leave", err)
	}
	if employee.ManagerID == nil {
		return nil, vo.PaginationDTOResponse{}, NewDomainError("Employee", fmt.Errorf("the employee you're requesting to view is not your staff"))
	}
	if *employee.ManagerID != manager.Id {
		return nil, vo.PaginationDTOResponse{}, NewDomainError("Employee", fmt.Errorf("the employee you're requesting to view is not your staff"))
	}

	attendances, page, err := uc.attRepo.GetMyAttendancesHistory(ctx, employeeId, q)
	if err != nil {
		return nil, page, NewRepositoryError("Leave", err)
	}

	return attendances, page, nil
}

/*
*********************************
ACTOR: HR
*********************************
*/
func (uc *attendanceUseCase) RetrieveEmployeesAttendanceHistory(ctx context.Context, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	if _, err := q.CommonQuery.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Attendance", err)
	}

	q.CommonQuery.Pagination.Order = "clock_in_at"
	attendances, page, err := uc.attRepo.GetEmployeesAttendanceHistory(ctx, q)
	if err != nil {
		return nil, page, NewRepositoryError("Attendance", err)
	}

	return attendances, page, nil
}

func (uc *attendanceUseCase) RetrieveAnEmployeeAttendances(ctx context.Context, employeeId string, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	if _, err := q.TimeQuery.Extract(); err != nil {
		return nil, vo.PaginationDTOResponse{}, NewClientError("Leave", err)
	}

	attendances, page, err := uc.attRepo.GetMyAttendancesHistory(ctx, employeeId, q)
	if err != nil {
		return nil, page, NewRepositoryError("Leave", err)
	}

	return attendances, page, nil
}

func (uc *attendanceUseCase) RetrieveEmployeesTodaysAttendances(ctx context.Context, q vo.HistoryAttendancesQuery) ([]entity.Attendance, vo.PaginationDTOResponse, error) {
	q.CommonQuery.Pagination.Order = "clock_in_at"
	attendances, page, err := uc.attRepo.GetEmployeesTodaysAttendances(ctx, q)
	if err != nil {
		return nil, page, NewRepositoryError("Attendance", err)
	}

	return attendances, page, nil
}

/*
*************************************************
MAILER HELPERS
*************************************************
*/
func (uc *attendanceUseCase) sendClockInOTPMail(receiverEmail string, data map[string]any) {
	uc.mailService.SendEmail(receiverEmail, service.OTP, data)
}

func (uc *attendanceUseCase) sendOvertimeSubmissionEmail(receiver entity.Employee, attendance entity.Attendance) {
	data := map[string]any{
		"ManagerName":   receiver.FullName,
		"Duration":      utils.SanitizeDuration(time.Duration(attendance.Overtime.Duration)),
		"Date":          attendance.ClockOutAt.In(utils.CURRENT_LOC).Format(time.DateOnly),
		"Reason":        attendance.Overtime.Reason,
		"RequesteeName": attendance.Employee.FullName,
	}

	uc.mailService.SendEmail(receiver.Email, service.OVERTIME_SUBMISSION, data)
}
