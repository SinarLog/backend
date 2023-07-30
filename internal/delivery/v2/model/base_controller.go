package model

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"sinarlog.com/internal/app/usecase"
	"sinarlog.com/internal/entity/vo"
	"sinarlog.com/internal/utils"
)

type BaseControllerV2 struct{}

const API_VERSION = "2.0"

type Data struct {
	ApiVersion string `json:"apiVersion,omitempty"`
	Status     string `json:"status,omitempty"`
	Data       any    `json:"data,omitempty"`
	Paging     any    `json:"paging,omitempty"`
}

type Error struct {
	ApiVersion string `json:"apiVersion,omitempty"`
	Error      any    `json:"error,omitempty"`
}

func (bc BaseControllerV2) jsonErrResponse(c *gin.Context, code int, err any) {
	c.AbortWithStatusJSON(code, Error{
		ApiVersion: API_VERSION,
		Error:      err,
	})
}

func (bc BaseControllerV2) Ok(c *gin.Context, obj ...any) {
	if len(obj) == 0 {
		c.JSON(http.StatusOK, Data{
			ApiVersion: API_VERSION,
			Status:     "OK",
		})
		return
	}
	c.JSON(http.StatusOK, Data{
		ApiVersion: API_VERSION,
		Status:     "OK",
		Data:       obj[0],
	})
}

// OkWithPage sends a response with a 200 http status code
// to the client. It sends the data with a paging information.
// It requires two data to be sent. If there are less than
// two data, it calls `Ok` internally.
func (bc BaseControllerV2) OkWithPage(c *gin.Context, obj ...any) {
	if len(obj) < 2 {
		bc.Ok(c, obj)
		return
	}
	c.JSON(http.StatusOK, Data{
		ApiVersion: API_VERSION,
		Status:     "OK",
		Data:       obj[0],
		Paging:     obj[1],
	})
}

func (bc BaseControllerV2) Created(c *gin.Context, obj ...any) {
	if len(obj) == 0 {
		c.JSON(http.StatusCreated, Data{
			ApiVersion: API_VERSION,
			Status:     "OK",
		})
		return
	}
	c.JSON(http.StatusCreated, Data{
		ApiVersion: API_VERSION,
		Status:     "OK",
		Data:       obj[0],
	})
}

func (bc BaseControllerV2) ClientError(c *gin.Context, err error) {
	bc.jsonErrResponse(c, http.StatusBadRequest, err)
}

func (bc BaseControllerV2) Unauthorized(c *gin.Context, err error) {
	bc.jsonErrResponse(c, http.StatusUnauthorized, err)
}

func (bc BaseControllerV2) NotFound(c *gin.Context, err error) {
	bc.jsonErrResponse(c, http.StatusNotFound, err)
}

func (bc BaseControllerV2) RequestTimeout(c *gin.Context, err error) {
	bc.jsonErrResponse(c, http.StatusRequestTimeout, err)
}

func (bc BaseControllerV2) Conflict(c *gin.Context, err error) {
	bc.jsonErrResponse(c, http.StatusConflict, err)
}

func (bc BaseControllerV2) UnprocessableEntity(c *gin.Context, err error) {
	bc.jsonErrResponse(c, http.StatusUnprocessableEntity, err)
}

func (bc BaseControllerV2) TooManyRequest(c *gin.Context) {
	bc.jsonErrResponse(c, http.StatusTooManyRequests, gin.H{"message": "Too many request, try again later"})
}

func (bc BaseControllerV2) UnexpectedError(c *gin.Context, err error) {
	bc.jsonErrResponse(c, http.StatusInternalServerError, err)
}

func (bc BaseControllerV2) SummariesUseCaseError(c *gin.Context, err any) {
	if appError, ok := err.(usecase.AppError); ok {
		switch appError.Type {
		case usecase.ErrUnexpected:
			bc.UnexpectedError(c, appError)
		case usecase.ErrRequestTimeout:
			bc.RequestTimeout(c, appError)
		case usecase.ErrInvalidInput:
			bc.ClientError(c, appError)
		case usecase.ErrUnprocessableEntity:
			bc.UnprocessableEntity(c, appError)
		case usecase.ErrBadRequest:
			bc.ClientError(c, appError)
		case usecase.ErrNotFound:
			bc.NotFound(c, appError)
		case usecase.ErrConflictState:
			bc.Conflict(c, appError)
		case usecase.ErrUnauthorized:
			bc.Unauthorized(c, appError)
		}
	} else {
		bc.UnexpectedError(c, usecase.AppError{
			Code:    500,
			Message: "unable to display message",
			Errors: []usecase.AppErrorDetail{
				{
					Message: "unable to display error message",
					Report:  "Please contact admin@example.com regarding this issue",
				},
			},
		})
	}
}

func (bc BaseControllerV2) ValidateImageFileHeader(header *multipart.FileHeader) error {
	// Validate extension
	split := strings.Split(header.Filename, ".")
	ext := split[len(split)-1]
	if err := validation.Validate(ext, validation.In("png", "jpeg", "jpg")); err != nil {
		return fmt.Errorf("file must be an image")
	}

	// Validate file size (compare in bytes)
	if header.Size > 2e+6 {
		return fmt.Errorf("file size is too large")
	}

	return nil
}

func (bc BaseControllerV2) ValidateAttachmentFileHeader(header *multipart.FileHeader) error {
	// Validate extension
	split := strings.Split(header.Filename, ".")
	ext := split[len(split)-1]
	if err := validation.Validate(ext, validation.In("png", "jpeg", "jpg", "pdf")); err != nil {
		return fmt.Errorf("file must be an image or a pdf")
	}

	// Validate file size (compare in bytes)
	if header.Size > 5e+6 {
		return fmt.Errorf("file size is too large")
	}

	return nil
}

// ParsePagination method    parses a pagination request into a VO.
// Keep in mind of these default values. Change in usecase if it
// doesn't meet the usecase requirements.
func (bc BaseControllerV2) ParsePagination(c *gin.Context) vo.PaginationDTORequest {
	page := c.DefaultQuery("page", "1")
	order := c.DefaultQuery("order", "created_at")
	sort := c.DefaultQuery("sort", "DESC")
	size := c.Query("size")

	return vo.PaginationDTORequest{
		Page:  page,
		Size:  size,
		Order: order,
		Sort:  sort,
	}
}

// ParseTimeQuery method    parses a time query request into a VO.
func (bc BaseControllerV2) ParseTimeQueryWithDefault(c *gin.Context) vo.TimeQueryDTORequest {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	month := c.DefaultQuery("month", fmt.Sprintf("%d", time.Now().In(utils.CURRENT_LOC).Month()))
	year := c.DefaultQuery("year", fmt.Sprintf("%d", time.Now().In(utils.CURRENT_LOC).Year()))

	return vo.TimeQueryDTORequest{
		StartDate: startDate,
		EndDate:   endDate,
		Month:     month,
		Year:      year,
	}
}

// ParseTimeQuery method    parses a time query request into a VO.
func (bc BaseControllerV2) ParseTimeQuery(c *gin.Context) vo.TimeQueryDTORequest {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	month := c.Query("month")
	year := c.Query("year")

	return vo.TimeQueryDTORequest{
		StartDate: startDate,
		EndDate:   endDate,
		Month:     month,
		Year:      year,
	}
}
