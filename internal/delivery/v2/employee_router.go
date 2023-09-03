package v2

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/delivery/v2/dto/mapper"
	"sinarlog.com/internal/delivery/v2/model"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type EmployeeController struct {
	model.BaseControllerV2
	attUC   usecase.IAttendanceUseCase
	leaveUC usecase.ILeaveUseCase
	emplUC  usecase.IEmployeeUseCase
	analUC  usecase.IAnalyticsUseCase
}

func NewEmployeeController(rg *gin.RouterGroup, attUC usecase.IAttendanceUseCase, leaveUC usecase.ILeaveUseCase, emplUC usecase.IEmployeeUseCase, analUC usecase.IAnalyticsUseCase) {
	controller := new(EmployeeController)
	controller.attUC = attUC
	controller.leaveUC = leaveUC
	controller.emplUC = emplUC
	controller.analUC = analUC

	att := rg.Group("/attendances")
	{
		att.GET("/active", controller.getTodaysAttendance)
		att.GET("/clockin", controller.requestClockInHandler)
		att.POST("/clockin", controller.clockInHandler)
		att.GET("/clockout", controller.requestClockOutHandler)
		att.POST("/clockout", controller.clockOutHandler)

		att.GET("/history", controller.getMyAttendancesLog)
	}

	leaves := rg.Group("/leaves")
	{
		leaves.GET("", controller.getMyLeaveRequestsHandler)
		leaves.GET("/quota", controller.getEmployeeLeaveQuotaHandler)
		leaves.GET("/:id", controller.getLeaveRequestById)
		leaves.POST("/report", controller.getLeaveRequestReportHandler)
		leaves.POST("", controller.applyForLeaveHandler)
	}

	ov := rg.Group("/overtimes")
	{
		ov.GET("", controller.getMyOvertimeSubmissions)
		ov.GET("/:id", controller.getMyOvertimeSubmissionDetailHandler)
	}

	empl := rg.Group("/employees")
	{
		empl.GET("", controller.getEmployeeListHandler)
		empl.GET(":id", controller.getEmployeeDetailHandler)
		empl.GET("/me/biodata", controller.getMyBiodataHandler)
		empl.GET("/whos-taking-leave", controller.whosTakingLeaveHandler)
	}

	anal := rg.Group("/anal")
	{
		anal.GET("/dashboard", controller.getDashboardAnalyticsHandler)
	}
}

func (controller *EmployeeController) requestClockInHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	if err := controller.attUC.RequestClockIn(c.Request.Context(), user); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}

func (controller *EmployeeController) clockInHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req dto.ClockInRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", err))
		return
	}

	vo := mapper.MapClockInRequestToVO(req)

	if err := controller.attUC.ClockIn(c.Request.Context(), user, vo); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Created(c)
}

func (controller *EmployeeController) requestClockOutHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	res, err := controller.attUC.RequestClockOut(c.Request.Context(), user)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapOvertimeOnAttendanceReportToResponse(res))
}

func (controller *EmployeeController) getTodaysAttendance(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	res, err := controller.attUC.RetrieveTodaysAttendance(c.Request.Context(), user)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapAttendanceEntityToResponse(res))
}

func (controller *EmployeeController) getMyAttendancesLog(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	user := c.Keys["user"].(entity.Employee)

	q := vo.HistoryAttendancesQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Early:  c.Query("early") == "true",
		Late:   c.Query("late") == "true",
		Closed: c.Query("closed") == "true",
	}

	res, page, err := controller.attUC.RetrieveMyAttendanceHistory(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapMyAttendanceLogToResponse(res), page)
}

func (controller *EmployeeController) clockOutHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req dto.ClockOutRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", err))
		return
	}

	if err := controller.attUC.ClockOut(c.Request.Context(), user, mapper.MapClockOutRequestToVO(req)); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Created(c)
}

func (controller *EmployeeController) getMyLeaveRequestsHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	user := c.Keys["user"].(entity.Employee)

	q := vo.LeaveProposalHistoryQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: c.Query("status"),
	}

	res, page, err := controller.leaveUC.RetrieveMyLeaves(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapMyLeaveRequestListToResponse(res), page)
}

func (controller *EmployeeController) getEmployeeLeaveQuotaHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	res, err := controller.leaveUC.RetrieveMyQuotas(c.Request.Context(), user)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapEmployeeLeaveQuotaToResponse(res))
}

func (controller *EmployeeController) getLeaveRequestReportHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req dto.LeaveRequest

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", err))
		return
	}

	leave, err := mapper.MapLeaveRequestToDomain(req)
	if err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", err))
		return
	}

	res, err := controller.leaveUC.RequestLeave(c.Request.Context(), user, leave)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapLeaveRequestReportToResponse(res))
}

func (controller *EmployeeController) applyForLeaveHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req dto.LeaveDecision

	// Process form body
	form := c.PostForm("leave")
	// Ignore if there are no files uploaded
	attachment, attachmentHeader, _ := c.Request.FormFile("attachment")
	// Parse json form body
	if err := json.Unmarshal([]byte(form), &req); err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", fmt.Errorf("body payload format is wrong")))
		return
	}

	// Checks file uploaded
	if attachmentHeader != nil {
		if err := controller.ValidateAttachmentFileHeader(attachmentHeader); err != nil {
			controller.ClientError(c, usecase.NewClientError("Body", err))
			return
		}
	}

	decision, err := mapper.MapLeaveDecisionToVO(req)
	if err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", err))
	}

	if err := controller.leaveUC.ApplyForLeave(c.Request.Context(), user, decision, attachment); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Created(c)
}

func (controller *EmployeeController) getMyBiodataHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	res, err := controller.emplUC.RetrieveEmployeeBiodata(c.Request.Context(), user.ID)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapEmployeeBiodataToResponse(res))
}

func (controller *EmployeeController) getEmployeeListHandler(c *gin.Context) {
	pagination := controller.ParsePagination(c)
	user := c.Keys["user"].(entity.Employee)
	fullName := c.Query("fullName")
	jobId := c.Query("jobId")

	q := vo.AllEmployeeQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: pagination,
		},
		FullName: fullName,
		JobId:    jobId,
	}

	res, page, err := controller.emplUC.RetrieveEmployeesList(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapEmployeeListToBriefEmployeeListResponse(res), page)
}

func (controller *EmployeeController) getEmployeeDetailHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	res, err := controller.emplUC.RetrieveEmployeeFullProfile(c.Request.Context(), user, c.Param("id"))
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapEmployeeFullProfileToResponse(res))
}

func (controller *EmployeeController) getLeaveRequestById(c *gin.Context) {
	leaveId := c.Param("id")
	res, err := controller.leaveUC.RetrieveLeaveRequest(c.Request.Context(), leaveId)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapMyLeaveRequestDetailToResponse(res))
}

func (controller *EmployeeController) getDashboardAnalyticsHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	res, err := controller.analUC.RetrieveDashboardAnalyticsForEmployeeById(c.Request.Context(), user.ID)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, res)
}

func (controller *EmployeeController) whosTakingLeaveHandler(c *gin.Context) {
	q := controller.ParseTimeQueryWithDefault(c)
	p := controller.ParsePagination(c)
	v := c.Query("version")

	query := vo.CommonQuery{
		TimeQuery:  q,
		Pagination: p,
	}

	if v == "mobile" {
		res, page, err := controller.leaveUC.RetrieveWhosTakingLeaveMobile(c.Request.Context(), query)
		if err != nil {
			controller.SummariesUseCaseError(c, err)
			return
		}
		controller.OkWithPage(c, mapper.MapIncomingLeaveProposalsForHrResponse(res), page)
	} else {
		res, err := controller.leaveUC.RetrieveWhosTakingLeave(c.Request.Context(), query)
		if err != nil {
			controller.SummariesUseCaseError(c, err)
			return
		}
		controller.Ok(c, res)
	}
}

func (controller *EmployeeController) getMyOvertimeSubmissions(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	status := c.DefaultQuery("status", "all")
	user := c.Keys["user"].(entity.Employee)

	q := vo.MyOvertimeSubmissionsQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: status,
	}

	overtimes, page, err := controller.attUC.RetrieveMyOvertimeSubmissions(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapMyOvertimeSubmissonToResponse(overtimes), page)
}

func (controller *EmployeeController) getMyOvertimeSubmissionDetailHandler(c *gin.Context) {
	res, err := controller.attUC.RetrieveOvertimeSubmission(c.Request.Context(), c.Param("id"))
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapOvertimeDetailToResponse(res))
}
