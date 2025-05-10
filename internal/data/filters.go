package data

import "github.com/meistens/api_practice/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string // holds supported sort values
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// check that the page and page_sixe params contain sensible values
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// check that the sort param matches a value in the safelist
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
