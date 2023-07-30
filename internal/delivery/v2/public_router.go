package v2

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/delivery/v2/dto/mapper"
	"sinarlog.com/internal/delivery/v2/model"
)

type PublicController struct {
	model.BaseControllerV2
	jobUC  usecase.IJobUseCase
	roleUC usecase.IRoleUseCase
	cfgUC  usecase.IConfigUseCase
	credUC usecase.ICredentialUseCase
}

func NewPublicController(
	rg *gin.RouterGroup,
	jobUC usecase.IJobUseCase,
	roleUC usecase.IRoleUseCase,
	cfgUC usecase.IConfigUseCase,
	credUC usecase.ICredentialUseCase,
) {
	controller := new(PublicController)
	controller.jobUC = jobUC
	controller.roleUC = roleUC
	controller.cfgUC = cfgUC
	controller.credUC = credUC

	rg.GET("/roles", controller.getRolesHandler)
	rg.GET("/jobs", controller.getJobsHandler)
	rg.GET("/configs", controller.getGlobalConfig)
	rg.POST("/forgot-password", controller.forgotPasswordHandler)
}

func (controller *PublicController) getRolesHandler(c *gin.Context) {
	data, err := controller.roleUC.RetrieveRoles(c.Request.Context())
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapRolesResponse(data))
}

func (controller *PublicController) getJobsHandler(c *gin.Context) {
	data, err := controller.jobUC.RetrieveJobs(c.Request.Context())
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapJobsResponse(data))
}

func (controller *PublicController) getGlobalConfig(c *gin.Context) {
	res, err := controller.cfgUC.RetrieveConfiguration(c.Request.Context())
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapConfigResponse(res))
}

func (controller *PublicController) forgotPasswordHandler(c *gin.Context) {
	var req dto.ForgotPassword

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.SummariesUseCaseError(c, usecase.NewClientError("Body", fmt.Errorf("error payload format, make sure you sent the right format")))
		return
	}

	if err := controller.credUC.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}
