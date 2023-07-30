package vo

import (
	"math"
	"strconv"
	"strings"

	"sinarlog.com/config"
)

// PaginationDTORequest struct    represents the struct of an incoming
// pagination request. This contains the uncleaned pagination.
type PaginationDTORequest struct {
	Page  string
	Size  string
	Order string
	Sort  string
}

// PaginationQuerySQL struct    represents a pagination query struct.
// It contains all the necessary queries.
type PaginationQuerySQL struct {
	// page stores the incoming requested page
	page    int
	Limit   int
	Offset  int
	OrderBy string
	Sort    string
}

// PaginationDTOResponse struct    represents the result of a paginated
// query.
type PaginationDTOResponse struct {
	Page        int `json:"page"`
	RowsPerPage int `json:"rowsPerPage"`
	TotalRows   int `json:"totalRows"`
	TotalPages  int `json:"totalPages"`
}

// PaginationInternalDTO struct    is used for transferring pagination
// objects through layers.
type PaginationInternalDTO struct {
	Page  int
	Size  int
	Order string
	Sort  string
}

// Extract method    extracts from a PaginationDTORequest to PaginationInternalDTO.
// During the extraction process, it fills  with default values then reads from
// the request.
func (p PaginationDTORequest) Extract() PaginationInternalDTO {
	cfg := config.GetConfig()

	pDTO := PaginationInternalDTO{
		Page:  1,
		Size:  cfg.App.DefaultPaginationSize,
		Order: p.Order,
		Sort:  "ASC",
	}

	page, err := strconv.Atoi(p.Page)
	if err == nil && page > 0 {
		pDTO.Page = page
	}

	size, err := strconv.Atoi(p.Size)
	if err == nil && size > 0 {
		pDTO.Size = size
	}

	if strings.ToUpper(p.Sort) == "ASC" || strings.ToUpper(p.Sort) == "DESC" {
		pDTO.Sort = p.Sort
	}

	return pDTO
}

// Extract method    extracts from a PaginationInternalDTO to PaginationQuerySQL.
// It calculates all the necessary numbers for a querly.
func (p PaginationInternalDTO) Extract() PaginationQuerySQL {
	pQuery := PaginationQuerySQL{
		page:    p.Page,
		OrderBy: p.Order,
		Sort:    p.Sort,
	}

	limit := p.Size
	offset := (p.Page - 1) * p.Size

	pQuery.Limit = limit
	pQuery.Offset = offset

	return pQuery
}

// MustExtract shortens the need for having to extract in each layer.
// YUnder the hood it calls Extract mthod in each layer of the pagibnation
// value objects.
func (p PaginationDTORequest) MustExtract() PaginationQuerySQL {
	return p.Extract().Extract()
}

// Compress method    summaries the pagination query result into a response for
// displaying to the client.
func (p PaginationQuerySQL) Compress(totalRows int64) PaginationDTOResponse {
	pRes := PaginationDTOResponse{}
	pRes.Page = p.page
	pRes.RowsPerPage = p.Limit
	pRes.TotalRows = int(totalRows)
	pRes.TotalPages = int(math.Ceil(float64(pRes.TotalRows) / float64(pRes.RowsPerPage)))

	return pRes
}
