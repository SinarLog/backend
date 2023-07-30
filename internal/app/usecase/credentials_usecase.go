package usecase

import (
	"context"
	"fmt"
	"log"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"sinarlog.com/internal/app/repo"
	"sinarlog.com/internal/app/service"
	"sinarlog.com/internal/entity"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

const (
	CLOCK_IN        = "CLOCK_IN"
	CHANGE_PASSWORD = "CHANGE_PASSWORD"
)

type credentialUseCase struct {
	repo          repo.ICredentialRepo
	service       service.IDoorkeeperService
	mailerService service.IMailerService
}

func NewCredentialUseCase(repo repo.ICredentialRepo, service service.IDoorkeeperService, mailerService service.IMailerService) ICredentialUseCase {
	return &credentialUseCase{repo: repo, service: service, mailerService: mailerService}
}

func (uc *credentialUseCase) Login(ctx context.Context, cred vo.Credential) (entity.Employee, vo.Credential, error) {
	if err := cred.ValidateAuthentication(); err != nil {
		return entity.Employee{}, vo.Credential{}, NewDomainError("Credentials", err)
	}

	employee, err := uc.repo.GetEmployeeByEmail(ctx, cred.Email)
	if err != nil {
		return entity.Employee{}, vo.Credential{}, NewNotFoundError("Credentials", err)
	}

	if err := uc.service.VerifyPassword(employee.Password, cred.Password); err != nil {
		return entity.Employee{}, vo.Credential{}, NewUnauthorizedError(fmt.Errorf("your password does not match"))
	}

	if employee.ResignedAt != nil || employee.Status == entity.RESIGNED {
		return entity.Employee{}, vo.Credential{}, NewUnauthorizedError(fmt.Errorf("you are no longer an employee of SinarLog"))
	}

	accessToken, err := uc.service.GenerateToken(employee)
	cred.AccessToken = accessToken

	return employee, cred, err
}

func (uc *credentialUseCase) Authorize(ctx context.Context, token string, roles ...any) (entity.Employee, error) {
	id, err := uc.service.VerifyAndParseToken(ctx, token)
	if err != nil {
		return entity.Employee{}, NewUnauthorizedError(err)
	}

	employee, err := uc.repo.GetEmployeeByIdV2(ctx, id)
	if err != nil {
		return entity.Employee{}, NewNotFoundError("Employee", err)
	}

	if err := validation.Validate(employee.Role.Code,
		validation.Required.Error("employee does not have a role"),
		validation.In(roles...).Error("roes does not match"),
	); err != nil {
		return entity.Employee{}, NewForbiddenError(err)
	}

	if employee.ResignedAt != nil || employee.Status == entity.RESIGNED {
		return entity.Employee{}, NewUnauthorizedError(fmt.Errorf("you are no longer an employee of SinarLog"))
	}

	return employee, nil
}

func (uc *credentialUseCase) ForgotPassword(ctx context.Context, email string) error {
	// Get the employee
	employee, err := uc.repo.GetEmployeeByEmail(ctx, email)
	if err != nil {
		return NewRepositoryError("Credential", err)
	}

	// Update its password
	password := utils.GenerateRandomPassword()
	hashedPassword, err := uc.service.HashPassword(password)
	if err != nil {
		return NewServiceError("Employee", err)
	}
	employee.Password = string(hashedPassword)

	// Persist
	if err := uc.repo.UpdateEmployeePassword(ctx, employee); err != nil {
		return NewRepositoryError("Crendetial", err)
	}

	// Send email
	go uc.sendForgotPasswordMail(employee.Email, map[string]any{
		"FullName": employee.FullName,
		"Email":    employee.Email,
		"Password": password,
	})

	return nil
}

func (uc *credentialUseCase) sendForgotPasswordMail(receiver string, data map[string]any) {
	if err := uc.mailerService.SendEmail(receiver, service.FORGOT_PASSWORD, data); err != nil {
		log.Printf("Unable to send forgot password email: %s", err)
	}
}
