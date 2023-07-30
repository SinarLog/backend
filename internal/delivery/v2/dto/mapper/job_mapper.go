package mapper

import (
	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/entity"
)

func MapJobsResponse(jobs []entity.Job) []dto.JobResponse {
	var res []dto.JobResponse

	for _, v := range jobs {
		res = append(res, dto.JobResponse{
			ID:   v.Id,
			Name: v.Name,
		})
	}

	return res
}
