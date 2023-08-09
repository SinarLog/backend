package composer

import (
	"sinarlog.com/internal/app/service"
	"sinarlog.com/internal/app/usecase"
)

type IUseCaseComposer interface {
	RaterUseCase() service.IRaterService
	CredentialUseCase() usecase.ICredentialUseCase
	ConfigUseCase() usecase.IConfigUseCase
	JobUseCase() usecase.IJobUseCase
	RoleUseCase() usecase.IRoleUseCase
	EmployeeUseCase() usecase.IEmployeeUseCase
	AttendanceUseCase() usecase.IAttendanceUseCase
	LeaveUseCase() usecase.ILeaveUseCase
	AnalyticsUseCase() usecase.IAnalyticsUseCase
	ChatUseCase() usecase.IChatUseCase
}

type useCaseComposer struct {
	repo    IRepoComposer
	service IServiceComposer
}

func NewUseCaseComposer(repo IRepoComposer, service IServiceComposer) IUseCaseComposer {
	return &useCaseComposer{repo: repo, service: service}
}

func (c *useCaseComposer) RaterUseCase() service.IRaterService {
	return c.service.RaterService()
}

func (c *useCaseComposer) CredentialUseCase() usecase.ICredentialUseCase {
	return usecase.NewCredentialUseCase(c.repo.CredentialRepo(), c.service.DoorkeeperService(), c.service.MailerService())
}

func (c *useCaseComposer) ConfigUseCase() usecase.IConfigUseCase {
	return usecase.NewConfigUseCase(c.repo.ConfigRepo())
}

func (c *useCaseComposer) EmployeeUseCase() usecase.IEmployeeUseCase {
	return usecase.NewEmployeeUseCase(
		c.repo.EmployeeRepo(),
		c.repo.ConfigRepo(),
		c.repo.SharedRepo(),
		c.repo.CredentialRepo(),
		c.service.DoorkeeperService(),
		c.service.MailerService(),
		c.service.BucketService(),
	)
}

func (c *useCaseComposer) JobUseCase() usecase.IJobUseCase {
	return usecase.NewJobUseCase(c.repo.JobRepo())
}

func (c *useCaseComposer) RoleUseCase() usecase.IRoleUseCase {
	return usecase.NewRoleUseCase(c.repo.RoleRepo())
}

func (c *useCaseComposer) AttendanceUseCase() usecase.IAttendanceUseCase {
	return usecase.NewAttendaceUseCase(
		c.repo.AttendanceRepo(),
		c.repo.LeaveRepo(),
		c.repo.ConfigRepo(),
		c.repo.EmployeeRepo(),
		c.service.DoorkeeperService(),
		c.service.MailerService(),
		c.service.NotifService(),
	)
}

func (c *useCaseComposer) LeaveUseCase() usecase.ILeaveUseCase {
	return usecase.NewLeaveUseCase(
		c.repo.LeaveRepo(),
		c.repo.EmployeeRepo(),
		c.repo.ConfigRepo(),
		c.service.MailerService(),
		c.service.NotifService(),
		c.service.BucketService(),
	)
}

func (c *useCaseComposer) AnalyticsUseCase() usecase.IAnalyticsUseCase {
	return usecase.NewAnalyticsUseCase(c.repo.AnalyticsRepo())
}

func (c *useCaseComposer) ChatUseCase() usecase.IChatUseCase {
	return usecase.NewChatUseCase(c.repo.ChatRepo(), c.repo.EmployeeRepo(), c.service.PubSubService())
}
