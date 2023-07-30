package vo

import (
	"fmt"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"sinarlog.com/internal/utils"
)

const shortForm = "2006-Jan-02"

type TimeQueryDTORequest struct {
	StartDate string
	EndDate   string
	Month     string
	Year      string
}

type TimeQueryInternalDTO struct {
	StartDate time.Time
	EndDate   time.Time
	Month     int
	Year      int
	Option    int
}

func (t TimeQueryDTORequest) validateTime() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Month,
			validation.Required.Error("please provide the month"),
			validation.Length(1, 2).Error("invalid month format"),
			is.Int.Error("month must be an integer"),
		),
		validation.Field(&t.Year,
			validation.Required.Error("please provide the year"),
			validation.Length(4, 4).Error("invalid year format"),
			is.Int.Error("year must be an integer"),
		),
	)
}

func (t TimeQueryInternalDTO) validateTime() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Month,
			validation.Required.Error("please provide the month"),
			validation.Min(1).Error("invalid month"),
			validation.Max(12).Error("invalid month"),
		),
		validation.Field(&t.Year,
			validation.Required.Error("please provide the year"),
			validation.Min(2000).Error("minimum year value is 2000"),
			validation.Max(time.Now().In(utils.CURRENT_LOC).Year()).Error("maximum year value is this year"),
		),
	)
}

func (t TimeQueryDTORequest) decide() int {
	switch {
	case t.StartDate != "" || t.EndDate != "":
		return 1
	case t.Month != "" || t.Year != "":
		return 2
	default:
		return 3
	}
}

func (t TimeQueryDTORequest) Extract() (TimeQueryInternalDTO, error) {
	switch t.decide() {
	case 1:
		sDate, err := time.Parse(shortForm, t.StartDate)
		if err != nil {
			return TimeQueryInternalDTO{}, fmt.Errorf("incorrect start date format... do \"2006-Jan-02\"")
		}
		eDate, err := time.Parse(shortForm, t.EndDate)
		if err != nil {
			return TimeQueryInternalDTO{}, fmt.Errorf("incorrect end date format... do \"2006-Jan-02\"")
		}

		compare := sDate.Compare(eDate)

		if compare == 1 || compare == 0 {
			return TimeQueryInternalDTO{}, fmt.Errorf("start date >= end date. Pick a different time")
		}
		qdto := TimeQueryInternalDTO{StartDate: sDate, EndDate: eDate, Option: 1}
		return qdto, nil
	case 2:
		if err := t.validateTime(); err != nil {
			return TimeQueryInternalDTO{}, err
		}
		month, _ := strconv.Atoi(t.Month)
		year, _ := strconv.Atoi(t.Year)
		qdto := TimeQueryInternalDTO{Month: month, Year: year, Option: 2}
		if err := qdto.validateTime(); err != nil {
			return TimeQueryInternalDTO{}, err
		}
		return qdto, nil
	default:
		return TimeQueryInternalDTO{Option: 3}, nil
	}
}
