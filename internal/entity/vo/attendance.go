package vo

import "sinarlog.com/internal/entity"

type ClockInRequest struct {
	Credential
	Loc entity.Point
}

type ClockOutPayload struct {
	Confirmation bool
	Reason       string
	Loc          entity.Point
}
