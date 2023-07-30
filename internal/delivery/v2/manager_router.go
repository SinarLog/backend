package v2

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/delivery/v2/dto/mapper"
	"sinarlog.com/internal/delivery/v2/model"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type ManagerController struct {
	model.BaseControllerV2
	leaveUC usecase.ILeaveUseCase
	attUC   usecase.IAttendanceUseCase
	analUC  usecase.IAnalyticsUseCase
	emplUC  usecase.IEmployeeUseCase
}

func NewManagerController(rg *gin.RouterGroup, leaveUC usecase.ILeaveUseCase, attUC usecase.IAttendanceUseCase, analUC usecase.IAnalyticsUseCase, emplUC usecase.IEmployeeUseCase) {
	controller := new(ManagerController)
	controller.leaveUC = leaveUC
	controller.attUC = attUC
	controller.analUC = analUC
	controller.emplUC = emplUC

	proposals := rg.Group("/proposals")
	{
		proposals.GET("/leaves/incoming", controller.seeIncomingLeaveProposalsHandler)
		proposals.GET("/leaves/incoming/:id", controller.seeIncomingLeaveProposalDetailHandler)
		proposals.PATCH("/leaves/incoming", controller.takeActionOnLeaveProposalHandler)

		proposals.GET("/leaves/history", controller.getLeaveProposalHistoryHandler)
		proposals.GET("/leaves/history/:id", controller.getLeaveProposalHistoryDetailHandler)

		proposals.GET("/overtimes/incoming", controller.seeIncomingOvertimeSubmissionsHandler)
		proposals.GET("/overtimes/incoming/:id", controller.seeIncomingOvertimeSubmissionDetailHandler)
		proposals.PATCH("/overtimes/incoming", controller.takeActionOnOvertimeSubmissionHandler)

		proposals.GET("/overtimes/history", controller.getOvertimeSubmissionHistoryHandler)
		proposals.GET("/overtimes/history/:id", controller.seeIncomingOvertimeSubmissionDetailHandler)

	}

	rg.GET("/attendances/history", controller.getStaffsAttendancesLog)

	employees := rg.Group("/employees")
	{
		employees.GET("", controller.getEmployeeListHandler)
		employees.GET(":id", controller.getEmployeeDetailHandler)
		employees.GET("/leaves/:employeeId", controller.getStaffsLeaveRequestHandler)
		employees.GET("/overtimes/:employeeId", controller.getStaffsOvertimeSubmissionsHandler)
		employees.GET("/attendances/:employeeId", controller.getStaffsAttendanceHandler)
	}

	anal := rg.Group("/anal")
	{
		anal.GET("/dashboard", controller.getDashboardAnalyticsHandler)
	}
}

func (controller *ManagerController) seeIncomingLeaveProposalsHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)
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

	res, page, err := controller.leaveUC.SeeIncomingLeaveProposalsForManager(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapIncomingLeaveProposalsForManagerResponse(res), page)
}

func (controller *ManagerController) seeIncomingLeaveProposalDetailHandler(c *gin.Context) {
	res, err := controller.leaveUC.RetrieveLeaveRequest(c.Request.Context(), c.Param("id"))
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapIncomingLeaveProposalDetailForManagerResponse(res))
}

func (controller *ManagerController) takeActionOnLeaveProposalHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req vo.LeaveAction
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", fmt.Errorf("missing required fields")))
		return
	}

	if err := controller.leaveUC.TakeActionOnLeaveProposalForManager(c.Request.Context(), user, req); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}

func (controller *ManagerController) getDashboardAnalyticsHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	res, err := controller.analUC.RetrieveDashboardAnalyticsForEmployeeById(c.Request.Context(), user.Id)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, res)
}

func (controller *ManagerController) seeIncomingOvertimeSubmissionsHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)
	p := controller.ParsePagination(c)
	n := c.Query("name")
	q := vo.IncomingOvertimeSubmissionsQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
		},
		Name: n,
	}

	res, page, err := controller.attUC.SeeIncomingOvertimeSubmissionsForManager(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapIncomingOvertimeSubmissionsToResponse(res), page)
}

func (controller *ManagerController) seeIncomingOvertimeSubmissionDetailHandler(c *gin.Context) {
	res, err := controller.attUC.RetrieveOvertimeSubmission(c.Request.Context(), c.Param("id"))
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapOvertimeDetailToResponse(res))
}

func (controller *ManagerController) takeActionOnOvertimeSubmissionHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req vo.OvertimeSubmissionAction
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.ClientError(c, usecase.NewClientError("Body", fmt.Errorf("missing required fields")))
		return
	}

	if err := controller.attUC.TakeActionOnOvertimeSubmissionByManager(c.Request.Context(), user, req); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}

func (controller *ManagerController) getLeaveProposalHistoryHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	s := c.DefaultQuery("status", "all")
	n := c.Query("name")
	user := c.Keys["user"].(entity.Employee)

	q := vo.LeaveProposalHistoryQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: s,
		Name:   n,
	}

	res, page, err := controller.leaveUC.RetrieveLeaveProposalsHistoryForManager(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapIncomingLeaveProposalsForManagerResponse(res), page)
}

func (controller *ManagerController) getOvertimeSubmissionHistoryHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	s := c.DefaultQuery("status", "all")
	n := c.Query("name")
	user := c.Keys["user"].(entity.Employee)

	q := vo.LeaveProposalHistoryQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: s,
		Name:   n,
	}

	res, page, err := controller.attUC.RetrieveOvertimeSubmissionsHistoryForManager(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapIncomingOvertimeSubmissionsToResponse(res), page)
}

func (controller *ManagerController) getLeaveProposalHistoryDetailHandler(c *gin.Context) {
	leaveId := c.Param("id")
	res, err := controller.leaveUC.RetrieveLeaveRequest(c.Request.Context(), leaveId)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapLeaveRequestDetailToResponse(res))
}

func (controller *ManagerController) getStaffsAttendancesLog(c *gin.Context) {
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
		Name:   c.Query("name"),
	}

	res, page, err := controller.attUC.RetrieveMyStaffsAttendanceHistory(c.Request.Context(), user, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapEmployeesAttendanceLogToResponse(res), page)
}

func (controller *ManagerController) getStaffsLeaveRequestHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	employeeId := c.Param("employeeId")
	user := c.Keys["user"].(entity.Employee)

	q := vo.LeaveProposalHistoryQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: c.Query("status"),
	}

	res, page, err := controller.leaveUC.RetrieveMyEmployeesLeaveHistory(c.Request.Context(), user, employeeId, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapMyLeaveRequestListToResponse(res), page)
}

func (controller *ManagerController) getStaffsOvertimeSubmissionsHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	employeeId := c.Param("employeeId")
	user := c.Keys["user"].(entity.Employee)

	q := vo.MyOvertimeSubmissionsQuery{
		CommonQuery: vo.CommonQuery{
			Pagination: p,
			TimeQuery:  t,
		},
		Status: c.Query("status"),
	}

	res, page, err := controller.attUC.RetrieveMyEmployeesOvertime(c.Request.Context(), user, employeeId, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapMyOvertimeSubmissonToResponse(res), page)
}

func (controller *ManagerController) getStaffsAttendanceHandler(c *gin.Context) {
	p := controller.ParsePagination(c)
	t := controller.ParseTimeQuery(c)
	user := c.Keys["user"].(entity.Employee)
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

	res, page, err := controller.attUC.RetrieveStaffAttendanceHistory(c.Request.Context(), user, employeeId, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapEmployeesAttendanceLogToResponse(res), page)
}

func (controller *ManagerController) getEmployeeListHandler(c *gin.Context) {
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

func (controller *ManagerController) getEmployeeDetailHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	res, err := controller.emplUC.RetrieveEmployeeFullProfile(c.Request.Context(), user, c.Param("id"))
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapEmployeeFullProfileToResponse(res))
}
