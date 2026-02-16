package services

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
)

type PropertyService struct {
	db *sql.DB
}

func NewPropertyService(db *sql.DB) *PropertyService {
	return &PropertyService{db}
}

type PropertyNameSuggestion struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Units []string `json:"units"`
}

type PropertyAutoCompleteResult struct {
	CompletionType string                   `json:"completionType"`
	Values         []string                 `json:"values"`
	Suggestions    []PropertyNameSuggestion `json:"suggestions"`
}

type PropertyAutoCompleteParams struct {
	UserId string
	Name   string
	Value  *string
}

func (ps *PropertyService) AutoComplete(ctx context.Context, params PropertyAutoCompleteParams) (*PropertyAutoCompleteResult, error) {
	userId, name, value := params.UserId, params.Name, params.Value

	sharedThingIds, err := operations.GetSharedThingIdsForUser(ctx, ps.db, userId)
	if err != nil {
		return nil, err
	}

	ownedThingIds, err := operations.GetOwnedThingIds(ctx, ps.db, userId)
	if err != nil {
		return nil, err
	}
	sharedThingIds = append(sharedThingIds, ownedThingIds...)

	if value == nil {
		return ps.autoCompleteName(ctx, name, sharedThingIds)
	}
	return ps.autoCompleteValue(ctx, name, *value, sharedThingIds)
}

type nameTypeUnitCount struct {
	Name  string      `boil:"name"`
	Type  string      `boil:"type"`
	Unit  null.String `boil:"unit"`
	Count int         `boil:"cnt"`
}

func (ps *PropertyService) autoCompleteName(ctx context.Context, name string, thingIds []string) (*PropertyAutoCompleteResult, error) {
	likeNameExpr := fmt.Sprintf("%s%%", name)

	var rows []nameTypeUnitCount
	err := models.NewQuery(
		qm.Select("name, type, unit, COUNT(*) as cnt"),
		qm.From("properties"),
		models.PropertyWhere.ThingID.IN(thingIds),
		models.PropertyWhere.Name.ILIKE(likeNameExpr),
		qm.GroupBy("name, type, unit"),
		qm.OrderBy("cnt DESC"),
	).Bind(ctx, ps.db, &rows)
	if err != nil {
		return nil, err
	}

	type nameAccum struct {
		typeCounts map[string]int
		totalCount int
		unitOrder  []string
		unitCounts map[string]int
	}
	nameMap := make(map[string]*nameAccum)
	nameOrder := []string{}

	for _, row := range rows {
		acc, ok := nameMap[row.Name]
		if !ok {
			acc = &nameAccum{
				typeCounts: make(map[string]int),
				unitCounts: make(map[string]int),
			}
			nameMap[row.Name] = acc
			nameOrder = append(nameOrder, row.Name)
		}
		acc.typeCounts[row.Type] += row.Count
		acc.totalCount += row.Count
		if row.Unit.Valid && row.Unit.String != "" {
			if _, seen := acc.unitCounts[row.Unit.String]; !seen {
				acc.unitOrder = append(acc.unitOrder, row.Unit.String)
			}
			acc.unitCounts[row.Unit.String] += row.Count
		}
	}

	// Sort names by total usage descending
	sort.Slice(nameOrder, func(i, j int) bool {
		return nameMap[nameOrder[i]].totalCount > nameMap[nameOrder[j]].totalCount
	})

	suggestions := make([]PropertyNameSuggestion, 0, len(nameOrder))
	for _, n := range nameOrder {
		acc := nameMap[n]

		dominantType := ""
		maxCount := 0
		for t, c := range acc.typeCounts {
			if c > maxCount {
				maxCount = c
				dominantType = t
			}
		}

		var units []string
		if dominantType == string(models.PropertyTypeFloat) && len(acc.unitOrder) > 0 {
			units = acc.unitOrder
		}

		suggestions = append(suggestions, PropertyNameSuggestion{
			Name:  n,
			Type:  dominantType,
			Units: units,
		})
	}

	return &PropertyAutoCompleteResult{
		CompletionType: "name",
		Suggestions:    suggestions,
	}, nil
}

func (ps *PropertyService) autoCompleteValue(ctx context.Context, name string, value string, thingIds []string) (*PropertyAutoCompleteResult, error) {
	likeValueExpr := fmt.Sprintf("%s%%", value)
	properties, err := models.Properties(
		models.PropertyWhere.ThingID.IN(thingIds),
		models.PropertyWhere.Name.EQ(name),
		models.PropertyWhere.ValueString.ILIKE(null.NewString(likeValueExpr, true)),
	).All(ctx, ps.db)
	if err != nil {
		return nil, err
	}

	resultSet := make(map[string]bool)
	for _, property := range properties {
		resultSet[property.ValueString.String] = true
	}
	result := []string{}
	for key := range resultSet {
		result = append(result, key)
	}
	return &PropertyAutoCompleteResult{
		CompletionType: "value",
		Values:         result,
	}, nil
}
