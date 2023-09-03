package mapper

import (
	"sinarlog.com/internal/delivery/v2/dto"
	"sinarlog.com/internal/entity"
)

func MapRolesResponse(roles []entity.Role) []dto.RoleResponse {
	var res []dto.RoleResponse

	for _, v := range roles {
		res = append(res, dto.RoleResponse{
			ID:   v.ID,
			Name: v.Name,
			Code: v.Code,
		})
	}

	return res
}
