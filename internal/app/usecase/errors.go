package usecase

import (
	"errors"
	"strings"
)

var (
	ErrUnexpected          = errors.New("something unexpected had occured")
	ErrRequestTimeout      = errors.New("query took too long or client cancelled the request")
	ErrInvalidInput        = errors.New("invalid input syntax")
	ErrUnprocessableEntity = errors.New("unable to process entity")
	ErrBadRequest          = errors.New("bad request")
	ErrNotFound            = errors.New("record not found")
	ErrConflictState       = errors.New("conflict state")
	ErrUnauthorized        = errors.New("unauthorized action")
	ErrForbidden           = errors.New("forbidden action")
)

var ErrNameMapper = map[error]string{
	ErrUnexpected:          "UnexpectedError",
	ErrRequestTimeout:      "RequestTimeOutError",
	ErrInvalidInput:        "InvalidInputError",
	ErrUnprocessableEntity: "DomainValidationError",
	ErrBadRequest:          "BadRequestError",
	ErrNotFound:            "NotFoundError",
	ErrConflictState:       "ConflictDuplicationError",
	ErrUnauthorized:        "UnauthorizedError",
	ErrForbidden:           "ForbiddenError",
}

type AppError struct {
	Code    int              `json:"code,omitempty"`
	Type    error            `json:"-"`
	Message string           `json:"message,omitempty"`
	Errors  []AppErrorDetail `json:"errors,omitempty"`
}

type AppErrorDetail struct {
	Domain  string `json:"domain,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
	Report  string `json:"report,omitempty"`
}

func (err AppError) Error() string {
	return err.Message
}

func NewError(domain string, code int, errType, err error) error {
	nErr := AppError{
		Code:    code,
		Type:    errType,
		Message: errType.Error(),
	}
	for _, message := range strings.Split(err.Error(), ";") {
		nErr.Errors = append(nErr.Errors, AppErrorDetail{
			Domain:  domain,
			Reason:  ErrNameMapper[errType],
			Message: strings.Trim(message, " "),
		})
	}

	return nErr
}

func NewErrorWithReport(domain string, code int, errType, err error, report string) error {
	nErr := AppError{
		Code:    code,
		Type:    errType,
		Message: errType.Error(),
	}
	for _, message := range strings.Split(err.Error(), ";") {
		nErr.Errors = append(nErr.Errors, AppErrorDetail{
			Domain:  domain,
			Reason:  ErrNameMapper[errType],
			Message: strings.Trim(message, " "),
			Report:  report,
		})
	}

	return nErr
}

func NewConflictError(domain string, err error) error {
	return NewError(domain, 409, ErrConflictState, err)
}

func NewDomainError(domain string, err error) error {
	return NewError(domain, 422, ErrUnprocessableEntity, err)
}

func NewRepositoryError(domain string, err error) error {
	return NewError(domain+" Repository", 500, ErrUnexpected, err)
}

func NewServiceError(domain string, err error) error {
	return NewError(domain+" Service", 500, ErrUnexpected, err)
}

func NewNotFoundError(domain string, err error) error {
	return NewError(domain, 404, ErrNotFound, err)
}

func NewUnauthorizedError(err error) error {
	return NewError("", 401, ErrUnauthorized, err)
}

func NewClientError(domain string, err error) error {
	return NewError(domain, 400, ErrBadRequest, err)
}

func NewForbiddenError(err error) error {
	return NewError("Credentials", 403, ErrForbidden, err)
}
