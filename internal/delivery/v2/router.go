package v2

import (
	"github.com/gin-gonic/gin"
	"sinarlog.com/internal/composer"
	"sinarlog.com/internal/delivery/middleware"
	"sinarlog.com/pkg/logger"
)

func NewRouter(r *gin.Engine, logger *logger.AppLogger, ucComposer composer.IUseCaseComposer) {
	r.Use(middleware.LogRequestMiddleware(logger))
	r.Use(middleware.CORSMiddleware())

	v2 := r.Group("/api/v2")
	{
		NewCredentialController(v2, ucComposer.CredentialUseCase(), ucComposer.RaterUseCase())

		ws := v2.Group("/ws")
		{
			NewWebsocketController(ws)
		}

		chat := v2.Group("/chat")
		{
			NewChatController(chat, ucComposer.CredentialUseCase(), ucComposer.EmployeeUseCase(), ucComposer.ChatUseCase())
		}

		pub := v2.Group("/pub")
		{
			NewPublicController(pub, ucComposer.JobUseCase(), ucComposer.RoleUseCase(), ucComposer.ConfigUseCase(), ucComposer.CredentialUseCase())
		}

		hr := v2.Group("/hr", middleware.NewMiddleware().AuthMiddleware(ucComposer.CredentialUseCase(), "hr"))
		{
			NewHrController(hr, ucComposer.EmployeeUseCase(), ucComposer.LeaveUseCase(), ucComposer.AttendanceUseCase(), ucComposer.ConfigUseCase(), ucComposer.AnalyticsUseCase())
		}

		mngr := v2.Group("/mngr", middleware.NewMiddleware().AuthMiddleware(ucComposer.CredentialUseCase(), "mngr"))
		{
			NewManagerController(mngr, ucComposer.LeaveUseCase(), ucComposer.AttendanceUseCase(), ucComposer.AnalyticsUseCase(), ucComposer.EmployeeUseCase())
		}

		empl := v2.Group("/empl", middleware.NewMiddleware().AuthMiddleware(ucComposer.CredentialUseCase(), "mngr", "staff"))
		{
			NewEmployeeController(empl, ucComposer.AttendanceUseCase(), ucComposer.LeaveUseCase(), ucComposer.EmployeeUseCase(), ucComposer.AnalyticsUseCase())
		}

		prfl := v2.Group("/profile", middleware.NewMiddleware().AuthMiddleware(ucComposer.CredentialUseCase(), "mngr", "staff", "hr"))
		{
			NewProfileController(prfl, ucComposer.EmployeeUseCase())
		}
	}
}
