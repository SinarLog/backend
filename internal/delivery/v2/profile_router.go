package v2

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/delivery/v2/dto/mapper"
	"sinarlog.com/internal/delivery/v2/model"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
)

type ProfileController struct {
	model.BaseControllerV2
	emplUC usecase.IEmployeeUseCase
}

func NewProfileController(rg *gin.RouterGroup, emplUC usecase.IEmployeeUseCase) {
	controller := new(ProfileController)
	controller.emplUC = emplUC

	rg.GET("", controller.getMyProfileHandler)
	rg.GET("/logs", controller.getMyChangesLog)
	rg.PATCH("", controller.updateProfileDataHandler)
	rg.PATCH("/update-password", controller.updatePasswordHandler)
	rg.PATCH("/update-avatar", controller.updateAvatarHandler)
}

func (controller *ProfileController) getMyProfileHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	res, err := controller.emplUC.RetrieveMyProfile(c.Request.Context(), user)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c, mapper.MapEmployeeFullProfileToResponse(res))
}

func (controller *ProfileController) updateProfileDataHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req vo.UpdateMyData

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.SummariesUseCaseError(c, usecase.NewClientError("Body", err))
		return
	}

	if err := controller.emplUC.UpdatePersonalData(c.Request.Context(), user, req); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}

func (controller *ProfileController) updatePasswordHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req vo.UpdatePassword

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		controller.SummariesUseCaseError(c, usecase.NewClientError("Body", err))
		return
	}

	if err := controller.emplUC.UpdatePassword(c.Request.Context(), user, req); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}

func (controller *ProfileController) updateAvatarHandler(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	var req dto.UpdateAvatar
	c.ShouldBind(&req)

	var avatar multipart.File
	if req.Avatar != nil {
		if err := controller.ValidateImageFileHeader(req.Avatar); err != nil {
			controller.ClientError(c, usecase.NewClientError("Body", err))
			return
		}

		a, err := req.Avatar.Open()
		if err != nil {
			controller.UnexpectedError(c, usecase.NewServiceError("File", err))
			return
		}
		avatar = a
	}

	if err := controller.emplUC.UpdateProfilePic(c.Request.Context(), user, avatar); err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.Ok(c)
}

func (controller *ProfileController) getMyChangesLog(c *gin.Context) {
	user := c.Keys["user"].(entity.Employee)

	q := vo.CommonQuery{
		Pagination: controller.ParsePagination(c),
	}

	res, page, err := controller.emplUC.RetrieveEmployeeChangesLog(c.Request.Context(), user.ID, q)
	if err != nil {
		controller.SummariesUseCaseError(c, err)
		return
	}

	controller.OkWithPage(c, mapper.MapEmployeeChangesLogToResponse(res), page)
}
