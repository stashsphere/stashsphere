package services

import (
	"context"
	"database/sql"
	"fmt"

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

type PropertyAutoCompleteResult struct {
	CompletionType string   `json:"completionType"`
	Values         []string `json:"values"`
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

	type ThingIdRow struct {
		ThingId string `boil:"thing_id"`
	}
	var thingIds []ThingIdRow
	err = models.NewQuery(
		qm.Distinct("id as thing_id"),
		qm.From("things"),
		qm.Where("owner_id = ?", userId),
	).Bind(ctx, ps.db, &thingIds)
	if err != nil {
		return nil, err
	}
	for _, thingIdRow := range thingIds {
		sharedThingIds = append(sharedThingIds, thingIdRow.ThingId)
	}

	var propertiesWhere []qm.QueryMod
	var completionType string
	if value == nil {
		likeNameExpr := fmt.Sprintf("%s%%", name)
		if err != nil {
			return nil, err
		}
		propertiesWhere = []qm.QueryMod{
			models.PropertyWhere.ThingID.IN(sharedThingIds),
			models.PropertyWhere.Name.ILIKE(likeNameExpr)}
		completionType = "name"
	} else {
		likeValueExpr := fmt.Sprintf("%s%%", *value)
		if err != nil {
			return nil, err
		}
		propertiesWhere = []qm.QueryMod{
			models.PropertyWhere.ThingID.IN(sharedThingIds),
			models.PropertyWhere.Name.EQ(name), models.PropertyWhere.ValueString.ILIKE(null.NewString(likeValueExpr, true))}
		completionType = "value"
	}
	properties, err := models.Properties(propertiesWhere...).All(ctx, ps.db)
	if err != nil {
		return nil, err
	}
	resultSet := make(map[string]bool)
	for _, property := range properties {
		if completionType == "value" {
			resultSet[property.ValueString.String] = true
		} else {
			resultSet[property.Name] = true
		}
	}
	result := []string{}
	for key, _ := range resultSet {
		result = append(result, key)
	}
	return &PropertyAutoCompleteResult{
		CompletionType: completionType,
		Values:         result,
	}, nil
}
