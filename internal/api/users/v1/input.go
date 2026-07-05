package usersv1

import (
	"math"
	"strings"

	"github.com/kitti12911/lib-util/v3/query"

	commonv1 "oas-sandbox/gen/grpc/common/v1"
)

// This file holds the HTTP-input adapters: query-string and JSON body values
// mapped onto proto request fields. Struct<->proto mappers, envelopes, and enum
// bridges are generated into mapper_generated.go by `mapgen map`.

func int32FromInt(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	return int32(value)
}

func paginationFromInput(input *ListUsersInput) *commonv1.PaginationRequest {
	return &commonv1.PaginationRequest{
		Page:     int32FromInt(input.Page),
		PageSize: int32FromInt(input.PageSize),
	}
}

// filterFromInput maps the flat GET query params onto a single leaf filter.
func filterFromInput(input *ListUsersInput) *commonv1.Filter {
	if input.FilterCol == "" {
		return nil
	}

	var vals []string
	if input.FilterVals != "" {
		vals = strings.Split(input.FilterVals, ",")
	}

	return &commonv1.Filter{
		Col:  input.FilterCol,
		Op:   query.FilterOpFromString[commonv1.FilterOp](input.FilterOp),
		Val:  input.FilterVal,
		Vals: vals,
	}
}

func orderByFromInput(input *ListUsersInput) []*commonv1.OrderBy {
	if input.OrderBy == "" {
		return nil
	}

	return []*commonv1.OrderBy{
		{
			Col:   input.OrderBy,
			Order: query.OrderDirectionFromString[commonv1.OrderDirection](input.Order),
		},
	}
}

func paginationFromAdvancedInput(input *AdvancedListUsersInput) *commonv1.PaginationRequest {
	if input.Body.Pagination == nil {
		return nil
	}

	return &commonv1.PaginationRequest{
		Page:     int32FromInt(input.Body.Pagination.Page),
		PageSize: int32FromInt(input.Body.Pagination.PageSize),
	}
}

func filterFromAdvancedInput(input *AdvancedListUsersInput) *commonv1.Filter {
	return filterToProto(input.Body.Filter)
}

// filterToProto converts a JSON filter tree to its proto form. A node with
// children is a group (col/op/val ignored); a leaf with an empty col is dropped,
// and a group left with no children collapses to nil.
func filterToProto(f *Filter) *commonv1.Filter {
	if f == nil {
		return nil
	}

	if len(f.Filters) > 0 {
		children := make([]*commonv1.Filter, 0, len(f.Filters))
		for i := range f.Filters {
			if child := filterToProto(&f.Filters[i]); child != nil {
				children = append(children, child)
			}
		}
		if len(children) == 0 {
			return nil
		}
		return &commonv1.Filter{
			Logic:   query.LogicalOpFromString[commonv1.LogicalOp](f.Logic),
			Filters: children,
		}
	}

	if f.Col == "" {
		return nil
	}
	return &commonv1.Filter{
		Col:  f.Col,
		Op:   query.FilterOpFromString[commonv1.FilterOp](f.Op),
		Val:  f.Val,
		Vals: f.Vals,
	}
}

func orderByFromAdvancedInput(input *AdvancedListUsersInput) []*commonv1.OrderBy {
	if len(input.Body.OrderBy) == 0 {
		return nil
	}

	orderBy := make([]*commonv1.OrderBy, 0, len(input.Body.OrderBy))
	for _, order := range input.Body.OrderBy {
		if order.Col == "" {
			continue
		}

		orderBy = append(orderBy, &commonv1.OrderBy{
			Col:   order.Col,
			Order: query.OrderDirectionFromString[commonv1.OrderDirection](order.Order),
		})
	}

	return orderBy
}
