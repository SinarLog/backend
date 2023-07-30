package v2

import (
	"fmt"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/delivery/middleware"
	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/delivery/v2/dto/mapper"
	"sinarlog.com/internal/delivery/v2/model"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type HrController struct {
	model.BaseControllerV2
	emplUC   usecase.IEmployeeUseCase
	leaveUC  usecase.ILeaveUseCase
	attUC    usecase.IAttendanceUseCase
	configUC usecase.IConfigUseCase
	analUC   usecase.IAnalyticsUseCase
}

func NewHrController(
	rg *gin.RouterGroup,
	emplUC usecase.IEmployeeUseCase,
	leaveUC usecase.ILeaveUseCase,
	attUC usecase.IAttendanceUseCase,
	configUC usecase.IConfigUseCase,
	analUC usecase.IAnalyticsUseCase,
) {
	controller := new(HrController)
	controller.emplUC = emplUC
	controller.leaveUC = leaveUC
	controller.attUC = attUC
	controller.configUC = configUC
	controller.analUC = analUC

	empl := rg.Group("/employees")
	{
		empl.GET("", controller.employeeListPagination(), controller.viewAllEmployeesHandler)
		empl.GET("/:id", controller.viewEmployeeFullProfile)
		empl.GET("/whos-taking-leave", controller.whosTakingLeaveHandler)
		empl.GET("/managers", controller.fetchManagersList)

		empl.POST("", controller.registerNewEmployeeHandler)
		empl.PATCH("/:id", controller.updateEmployeeDataHandler)

		empl.GET("/leaves/:employeeId", controller.getStaffEmployeeLeavesHandler)
		empl.GET("/overtimes/:employeeId", controller.getStaffsOvertimesHandler)
		empl.GET("/attendances/:employeeId", controller.getStaffAttendancesHandler)
		empl.GET("/logs/:employeeId", controller.getEmployeeDataChangesLogHandler)
	}

	proposals := rg.Group("/proposals")
	{
		proposals.GET("/leaves/incoming", controller.seeIncomingLeaveProposalsHandler)
		proposals.GET("/leaves/incoming/:id", controller.seeIncomingLeaveProposalDetailHandler)
		proposals.PATCH("/leaves/incoming", controller.takeActionOnLeaveProposalHandler)

		proposals.GET("/leaves/history", controller.getLeaveProposalHistoryHandler)
		proposals.GET("/leaves/history/:id", controller.getLeaveProposalHistoryDetailHandler)

		proposals.GET("/overtimes/history", controller.getOvertimeSubmissionHistoryHandler)
		proposals.GET("/overtimes/history/:id", controller.getOvertimeSubmissionHistoryDetailHandler)
	}

	attendances := rg.Group("/attendances")
	{
		attendances.GET("/history", controller.getEmployeesAttendancesLog)
		attendances.GET("/today", controller.getEmployeesTodaysAttendances)
	}

	cfg := rg.Group("/config")
	{
		cfg.GET("", controller.getConfigHandler)
		cfg.GET("/logs", controller.getChangesLogsHandler)
		cfg.PUT("", controller.updateConfigHandler)
	}

	anal := rg.Group("/anal")
	{
		anal.GET("", controller.getDashboardHrAnalyticsHandler)
	}
}

/*
--------------- Specific Endpoint Group Middlewares ---------------
*/
func (controller *HrController) employeeListPagination() gin.HandlerFunc {
	return middleware.NewMiddleware().PaginateMiddleware("join_date")
}

/*
--------------- Specific Endpoint Group Controllers ---------------
*/
func (controller *HrController) registerNewEmployeeHandler(c *gin.Context) {
	creator := c.Keys["user"].(entity.Employee)

	var req dto.CreateNewEmployeeRequest
	if err := c.ShouldBind(&req); err != nil {
		controller.ClientError(c, usecase.NewClientError("Form", err))
		return
	}

	payload, err := mapper.MapCreateNewEmployeeRequestToEmployeeEntity(req)
	if err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", err))
		return
	}

	var avatar multipart.File

	if req.Avatar != nil {
		// If avatar is sent, validate the file
		if err := controller.ValidateImageFileHeader(req.Avatar); err != nil {
			controller.ClientError(c, usecase.NewClientError("Body", err))
			return
		}

		avatar, err = req.Avatar.Open()
		if err != nil {
			controller.UnexpectedError(c, usecase.NewServiceError("File", err))
			return
		}
	}

	if err := controller.emplUC.RegisterNewEmployee(c.Request.Context(), creator, payload, avatar); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Created(c)
}

func (controller *HrController) viewAllEmployeesHandler(c *gin.Context) {
	pagination := c.Keys["pagination"].(vo.PaginationDTORequest)
	fullName := c.Query("fullName")
	jobId := c.Query("jobId")
	user := c.Keys["user"].(entity.Employee)

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

	data := mapper.MapEmployeeListToBriefEmployeeListResponse(res)

	controller.OkWithPage(c, data, page)
}

func (controller *HrController) fetchManagersList(c *gin.Context) {
	res, err := controller.emplUC.ViewManagersList(c.Request.Context())
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	data := mapper.MapManagersListToResponse(res)

	controller.Ok(c, data)
}

func (controller *HrController) viewEmployeeFullProfile(c *gin.Context) {
	employeeId := c.Param("id")
	user := c.Keys["user"].(entity.Employee)

	if employeeId == "" {
		controller.UnexpectedError(c, usecase.NewNotFoundError("Employee", fmt.Errorf("employee id is not found in request")))
		return
	}

	res, err := controller.emplUC.RetrieveEmployeeFullProfile(c.Request.Context(), user, employeeId)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapEmployeeFullProfileToResponse(res))
}

func (controller *HrController) seeIncomingLeaveProposalsHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	n := c.Query("name")
	q := vo.IncomingLeaveProposals{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Name: n,
	}

	res, page, err := controller.leaveUC.SeeIncomingLeaveProposalsForHr(c.Request.Context(), q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapIncomingLeaveProposalsForHrResponse(res), page)
}

func (controller *HrController) seeIncomingLeaveProposalDetailHandler(c *gin.Context) {
	res, err := controller.leaveUC.RetrieveLeaveRequest(c.Request.Context(), c.Param("id"))
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapIncomingLeaveProposalDetailForHrResponse(res))
}

func (controller *HrController) takeActionOnLeaveProposalHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req vo.LeaveAction
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", fmt.Errorf("missing required fields")))
		return
	}

	if err := controller.leaveUC.TakeActionOnLeaveProposalForHr(c.Request.Context(), user, req); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}

func (controller *HrController) getLeaveProposalHistoryHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	s := c.DefaultQuery("status", "all")
	n := c.Query("name")

	q := vo.LeaveProposalHistoryQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: s,
		Name:   n,
	}

	res, page, err := controller.leaveUC.RetrieveLeaveProposalsHistoryForHr(c.Request.Context(), q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapIncomingLeaveProposalsForManagerResponse(res), page)
}

func (controller *HrController) getLeaveProposalHistoryDetailHandler(c *gin.Context) {
	leaveId := c.Param("id")
	res, err := controller.leaveUC.RetrieveLeaveRequest(c.Request.Context(), leaveId)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapLeaveRequestDetailToResponse(res))
}

func (controller *HrController) whosTakingLeaveHandler(c *gin.Context) {
	q := controller.ParseTimeQueryWithDefault(c)

	query := vo.CommonQuery{
		TimeQuery: q,
	}
	res, err := controller.leaveUC.RetrieveWhosTakingLeave(c.Request.Context(), query)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}
	controller.Ok(c, res)
}

func (controller *HrController) getOvertimeSubmissionHistoryHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	s := c.DefaultQuery("status", "all")
	n := c.Query("name")

	q := vo.LeaveProposalHistoryQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: s,
		Name:   n,
	}

	res, page, err := controller.attUC.RetrieveOvertimeSubmissionsHistoryForHr(c.Request.Context(), q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapIncomingOvertimeSubmissionsToResponse(res), page)
}

func (controller *HrController) getOvertimeSubmissionHistoryDetailHandler(c *gin.Context) {
	res, err := controller.attUC.RetrieveOvertimeSubmission(c.Request.Context(), c.Param("id"))
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapOvertimeDetailToResponse(res))
}

func (controller *HrController) getEmployeesAttendancesLog(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)

	q := vo.HistoryAttendancesQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Early:  c.Query("early") == "true",
		Late:   c.Query("late") == "true",
		Closed: c.Query("closed") == "true",
		Name:   c.Query("name"),
	}

	res, page, err := controller.attUC.RetrieveEmployeesAttendanceHistory(c.Request.Context(), q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapEmployeesAttendanceLogToResponse(res), page)
}

func (controller *HrController) getEmployeesTodaysAttendances(c *gin.Context) {
	p := controller.ParsePagination(c)

	q := vo.HistoryAttendancesQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
		},
		Early:  c.Query("early") == "true",
		Late:   c.Query("late") == "true",
		Closed: c.Query("closed") == "true",
		Name:   c.Query("name"),
	}

	res, page, err := controller.attUC.RetrieveEmployeesTodaysAttendances(c.Request.Context(), q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapEmployeesAttendanceLogToResponse(res), page)
}

func (controller *HrController) getStaffEmployeeLeavesHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	employeeId := c.Param("employeeId")

	q := vo.LeaveProposalHistoryQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: c.Query("status"),
	}

	res, page, err := controller.leaveUC.RetrieveAnEmployeeLeaves(c.Request.Context(), employeeId, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapMyLeaveRequestListToResponse(res), page)
}

func (controller *HrController) getStaffsOvertimesHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	employeeId := c.Param("employeeId")

	q := vo.MyOvertimeSubmissionsQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: c.Query("status"),
	}

	res, page, err := controller.attUC.RetrieveAnEmployeeOvertimes(c.Request.Context(), employeeId, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapMyOvertimeSubmissonToResponse(res), page)
}

func (controller *HrController) getStaffAttendancesHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	employeeId := c.Param("employeeId")

	q := vo.HistoryAttendancesQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Early:  c.Query("early") == "true",
		Late:   c.Query("late") == "true",
		Closed: c.Query("closed") == "true",
		Name:   c.Query("name"),
	}

	res, page, err := controller.attUC.RetrieveAnEmployeeAttendances(c.Request.Context(), employeeId, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapEmployeesAttendanceLogToResponse(res), page)
}

func (controller *HrController) updateEmployeeDataHandler(c *gin.Context) {
	var payload vo.UpdateEmployeeData
	user := c.Keys["user"].(entity.Employee)

	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err != nil {
		controller.SummariesUseCaseError(c, usecase.NewClientError("Body", err))
		return
	}

	if err := controller.emplUC.UpdateEmployeeData(c.Request.Context(), user, c.Param("id"), payload); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}

func (controller *HrController) getEmployeeDataChangesLogHandler(c *gin.Context) {
	q := vo.CommonQuery{
		Pagination: controller.ParsePagination(c),
	}

	res, page, err := controller.emplUC.RetrieveEmployeeChangesLog(c.Request.Context(), c.Param("employeeId"), q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapEmployeeChangesLogToResponse(res), page)
}

func (controller *HrController) updateConfigHandler(c *gin.Context) {
	var payload dto.UpdateConfigRequest
	user := c.Keys["user"].(entity.Employee)

	if err := c.ShouldBindBodyWith(&payload, binding.JSON); err != nil {
		controller.SummariesUseCaseError(c, usecase.NewClientError("Body?", err))
		return
	}

	if err := controller.configUC.ChangeCompanyConfig(c.Request.Context(), user, mapper.MapUpdateConfigToDomain(payload)); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}

func (controller *HrController) getConfigHandler(c *gin.Context) {
	config, err := controller.configUC.RetrieveConfiguration(c.Request.Context())
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapConfigDetailResponse(config))
}

func (controller *HrController) getChangesLogsHandler(c *gin.Context) {
	q := vo.CommonQuery{
		Pagination: controller.ParsePagination(c),
	}

	res, page, err := controller.configUC.RetrieveChangesLogs(c.Request.Context(), q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapConfigChangesLogToResponse(res), page)
}

func (controller *HrController) getDashboardHrAnalyticsHandler(c *gin.Context) {
	anal, err := controller.analUC.RetrieveDashboardAnalyticsHr(c.Request.Context())
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, anal)
}
