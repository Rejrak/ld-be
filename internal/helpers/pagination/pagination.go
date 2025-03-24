package pagination

import (
	"be/internal/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

// PaginationResult represents the paginated result of a query.
// It includes information such as total results, page count, and data for the current page.
type PaginationResult struct {
	Total       int         `json:"total"`
	TotalPages  int         `json:"total_pages"`
	CurrentPage int         `json:"current_page"`
	Limit       int         `json:"limit"`
	Page        int         `json:"page"`
	Offset      int         `json:"offset"`
	Data        interface{} `json:"data"`
}

type FilterInput struct {
	Field string  `json:"field"`
	Value *string `json:"value,omitempty"`
}

type PaginationInput struct {
	Limit    int     `json:"limit"`
	Page     int     `json:"page"`
	OrderBy  *string `json:"orderBy,omitempty"`
	OrderDir *string `json:"orderDir,omitempty"`
}

// ApplyFilters applies filters to a GORM database query.
// Takes a GORM database instance, an array of filters, and an optional table prefix.
// Validates each filter and applies it to the query, returning an updated *gorm.DB instance or an error.
func ApplyFilters(db *gorm.DB, filters []*FilterInput, tablePrefix *string) (*gorm.DB, error) {

	for _, filter := range filters {
		field := utils.Data.CamelCaseToSnakeCase(filter.Field)

		if tablePrefix != nil {
			field = *tablePrefix + "." + field
		}

		if filter.Value != nil {
			isID := strings.Contains(strings.ToLower(filter.Field), "id")
			isDate := strings.Contains(strings.ToLower(filter.Field), "date")
			if isID {
				if !isValidUUID(*filter.Value) {
					return nil, fmt.Errorf("ID non valido: %s", *filter.Value)
				}
				db = db.Where(field+" = ?", filter.Value)
				continue
			} else if isDate {
				if strings.Contains(*filter.Value, "|") {
					if !isValidDateRange(*filter.Value) {
						return nil, fmt.Errorf("intervallo di date non valido: %s", *filter.Value)
					}
				} else {
					if !isValidDate(*filter.Value) {
						return nil, fmt.Errorf("data non valida: %s", *filter.Value)
					}
				}

				dateRange := strings.Split(*filter.Value, "|")
				if len(dateRange) == 2 {
					startDate := dateRange[0]
					endDate := dateRange[1]
					db = db.Where(field+" BETWEEN ? AND ?", startDate, endDate)
				} else {
					stdt := strings.Contains(field, "start")
					if stdt {
						db = db.Where(field+" >= ?", filter.Value)
					} else {
						db = db.Where(field+" <= ?", filter.Value)
					}
				}
				continue
			} else {
				db = db.Where(field+" LIKE ?", "%"+*filter.Value+"%")
			}
		}
	}
	return db, nil
}

// PaginateAndFilter applies pagination and filters to a database query and returns paginated results.
// Accepts a GORM query, model, pagination parameters, and filters, returning paginated data.
// Sets limits, applies filters, and manages sorting and offsets within the query.
func PaginateAndFilter[T any](ctx context.Context, db *gorm.DB, model T, params PaginationInput, filters []*FilterInput) (*PaginationResult, error) {
	if params.Limit <= 0 {
		params.Limit = 1
	}
	var total int64
	query := db.Model(&model)

	query, err := ApplyFilters(query, filters, nil)
	if err != nil {
		return nil, err
	}

	totalChan := make(chan error)
	defer close(totalChan)
	go func() {
		totalChan <- query.WithContext(ctx).Count(&total).Error
	}()

	if params.OrderBy != nil && params.OrderDir != nil {
		orderBy := utils.Data.CamelCaseToSnakeCase(*params.OrderBy)
		query = query.Order(orderBy + " " + *params.OrderDir)
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * params.Limit

	errTotalChan := <-totalChan
	if err := query.WithContext(ctx).Limit(params.Limit).Offset(offset).Find(&model).Error; err != nil || errTotalChan != nil {
		utils.Log.Error(ctx, params, err)
		return nil, errors.New("DB Communication error")
	}

	totalPages := int((total + int64(params.Limit) - 1) / int64(params.Limit))
	currentPage := page

	result := &PaginationResult{
		Total:       int(total),
		TotalPages:  totalPages,
		CurrentPage: currentPage,
		Limit:       params.Limit,
		Offset:      offset,
		Data:        model,
	}

	return result, nil
}

// isValidDate checks if a string represents a valid date in one of the specified formats.
// Attempts parsing the date using different formats and returns true if successful.
func isValidDate(dateStr string) bool {
	formats := []string{"2006-01-02", "2006-01-02 15:04:05", "02-01-2006"}
	for _, format := range formats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return true
		}
	}
	return false
}

// isValidUUID checks if a string represents a valid UUID.
// Uses the uuid package to parse and verify validity.
func isValidUUID(idStr string) bool {
	_, err := uuid.Parse(idStr)
	return err == nil
}

// isValidDateRange checks if a string represents a valid date range.
// The range must contain exactly two valid dates separated by the '|' character.
func isValidDateRange(dateRange string) bool {
	dates := strings.Split(dateRange, "|")
	if len(dates) != 2 {
		return false
	}
	return isValidDate(dates[0]) && isValidDate(dates[1])
}

// IsValidInput checks if the input fields are valid by comparing them with the model fields.
// Returns an error if any invalid field is found in the filter list.
func IsValidInput[T any](md T, filters []*FilterInput) error {
	controlValue := utils.Data.GetModelFields(md)
	controlValue = append(controlValue, "id")
	var validMap = make(map[string]bool)

	for _, field := range controlValue {
		validMap[field] = true
	}
	for _, filter := range filters {
		f := utils.Data.CamelCaseToSnakeCase(filter.Field)
		if _, ok := validMap[f]; !ok {
			return fmt.Errorf("invalid field: %s", filter.Field)
		}
	}
	return nil
}
