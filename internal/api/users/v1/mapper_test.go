package usersv1

import (
	"math"
	"testing"

	commonv1 "oas-sandbox/gen/grpc/common/v1"

	"github.com/stretchr/testify/assert"
)

// Only the hand-written input/filter helpers in input.go are tested here; the
// generated proto<->model mappers in mapper_generated.go are the generator's
// responsibility (covered by lib-orm/internal/mappergen) and are excluded from
// coverage in go-test.sh.

func TestInt32FromInt(t *testing.T) {
	t.Parallel()
	assert.Equal(t, int32(0), int32FromInt(0))
	assert.Equal(t, int32(42), int32FromInt(42))
	assert.Equal(t, int32(math.MaxInt32), int32FromInt(math.MaxInt32+1))
	assert.Equal(t, int32(math.MinInt32), int32FromInt(math.MinInt32-1))
}

func TestPaginationFromInput(t *testing.T) {
	t.Parallel()
	got := paginationFromInput(&ListUsersInput{Page: 2, PageSize: 25})
	assert.Equal(t, int32(2), got.Page)
	assert.Equal(t, int32(25), got.PageSize)
}

func TestFilterFromInputEmpty(t *testing.T) {
	t.Parallel()
	assert.Nil(t, filterFromInput(&ListUsersInput{}))
}

func TestFilterFromInputSingleValue(t *testing.T) {
	t.Parallel()
	got := filterFromInput(&ListUsersInput{
		FilterCol: "username", FilterOp: "like_ci", FilterVal: "kit",
	})
	assert.Equal(t, "username", got.Col)
	assert.Equal(t, "kit", got.Val)
	assert.Empty(t, got.Vals)
	assert.Empty(t, got.Filters)
}

func TestFilterFromInputMultipleValues(t *testing.T) {
	t.Parallel()
	got := filterFromInput(&ListUsersInput{
		FilterCol: "status", FilterOp: "in", FilterVals: "active,pending",
	})
	assert.Equal(t, []string{"active", "pending"}, got.Vals)
}

func TestOrderByFromInput(t *testing.T) {
	t.Parallel()
	assert.Nil(t, orderByFromInput(&ListUsersInput{}))
	got := orderByFromInput(&ListUsersInput{OrderBy: "username", Order: "desc"})
	assert.Len(t, got, 1)
	assert.Equal(t, "username", got[0].Col)
}

func TestPaginationFromAdvancedInput(t *testing.T) {
	t.Parallel()
	assert.Nil(t, paginationFromAdvancedInput(&AdvancedListUsersInput{}))

	got := paginationFromAdvancedInput(&AdvancedListUsersInput{
		Body: AdvancedListUsersRequest{Pagination: &Pagination{Page: 1, PageSize: 50}},
	})
	assert.Equal(t, int32(1), got.Page)
	assert.Equal(t, int32(50), got.PageSize)
}

func TestFilterFromAdvancedInput(t *testing.T) {
	t.Parallel()
	assert.Nil(t, filterFromAdvancedInput(&AdvancedListUsersInput{}))

	// leaf
	got := filterFromAdvancedInput(&AdvancedListUsersInput{
		Body: AdvancedListUsersRequest{
			Filter: &Filter{Col: "username", Op: "like_ci", Val: "kit"},
		},
	})
	assert.Equal(t, "username", got.Col)
	assert.Empty(t, got.Filters)
}

func TestFilterToProtoGroup(t *testing.T) {
	t.Parallel()
	// age >= 18 AND (status = active OR status = pending)
	got := filterToProto(&Filter{
		Logic: "and",
		Filters: []Filter{
			{Col: "age", Op: "gte", Val: "18"},
			{
				Logic: "or",
				Filters: []Filter{
					{Col: "status", Op: "exact", Val: "active"},
					{Col: "status", Op: "exact", Val: "pending"},
				},
			},
		},
	})
	assert.Equal(t, commonv1.LogicalOp_LOGICAL_OP_AND, got.Logic)
	assert.Len(t, got.Filters, 2)
	assert.Equal(t, "age", got.Filters[0].Col)
	assert.Equal(t, commonv1.LogicalOp_LOGICAL_OP_OR, got.Filters[1].Logic)
	assert.Len(t, got.Filters[1].Filters, 2)
}

func TestFilterToProtoDropsEmptyNodes(t *testing.T) {
	t.Parallel()
	// empty-col leaves are dropped; a group with no surviving children collapses
	assert.Nil(t, filterToProto(&Filter{}))
	assert.Nil(t, filterToProto(&Filter{
		Logic:   "or",
		Filters: []Filter{{Col: "", Val: "skip-me"}},
	}))

	got := filterToProto(&Filter{
		Logic: "or",
		Filters: []Filter{
			{Col: "username", Op: "like_ci", Val: "kit"},
			{Col: "", Op: "exact", Val: "skip-me"},
		},
	})
	assert.Len(t, got.Filters, 1)
	assert.Equal(t, "username", got.Filters[0].Col)
}

func TestOrderByFromAdvancedInput(t *testing.T) {
	t.Parallel()
	assert.Nil(t, orderByFromAdvancedInput(&AdvancedListUsersInput{}))

	got := orderByFromAdvancedInput(&AdvancedListUsersInput{
		Body: AdvancedListUsersRequest{
			OrderBy: []OrderBy{
				{Col: "username", Order: "asc"},
				{Col: "", Order: "desc"}, // skipped
			},
		},
	})
	assert.Len(t, got, 1)
	assert.Equal(t, "username", got[0].Col)
}
