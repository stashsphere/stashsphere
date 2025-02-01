package services

import (
	"database/sql"
)

type PropertyService struct {
	db *sql.DB
}

func NewPropertyService(db *sql.DB) *PropertyService {
	return &PropertyService{db}
}
