package data

import (
	"strings"

	"slices"

	"github.com/meistens/api_practice/internal/validator"
)

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

// checks that the client-provided Sort field matches one of the entries in the safelist[]
// if it does, extract the column name from the Sort field by stripping
// the leading hyphen character (if one exists)
func (f Filters) sortColumn() string {
	if slices.Contains(f.SortSafelist, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-")
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// return the sort direction (asc or desc) depending on the prefix
// character of the sort field
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}
