package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"sinarlog.com/internal/app/service"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/delivery/middleware"
	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/delivery/v2/dto/mapper"
	"sinarlog.com/internal/delivery/v2/model"
)

type CredentialController struct {
	model.BaseControllerV2
	uc  usecase.ICredentialUseCase
	srv service.IRaterService
}

func NewCredentialController(rg *gin.RouterGroup, uc usecase.ICredentialUseCase, srv service.IRaterService) {
	controller := new(CredentialController)
	controller.uc = uc
	controller.srv = srv

	r := rg.Group("/credentials")
	{
		// Rate limiting middleware
		r.Use(middleware.NewMiddleware().RateLimiterMiddleware(srv))
		r.POST("/login", controller.loginHandler)
	}
}

/*
*************************************************
Controllers
*************************************************
*/
func (controller *CredentialController) loginHandler(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.ClientError(c, err)
		return
	}

	cred := mapper.MapLoginRequestToCredentialVO(req)

	employee, cred, err := controller.uc.Login(c.Request.Context(), cred)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapToLoginResponse(employee, cred))
}
