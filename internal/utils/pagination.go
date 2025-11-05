package utils

import (
	"math"

	"gorm.io/gorm"
)

// Pagination define los parámetros y resultados comunes de una paginación.
type Pagination struct {
	Page       int         `json:"page"`        // Página actual
	Limit      int         `json:"limit"`       // Tamaño de página
	TotalRows  int64       `json:"total_rows"`  // Total de registros
	TotalPages int         `json:"total_pages"` // Total de páginas
	Data       interface{} `json:"data"`        // Resultado de la consulta
}

// Paginate aplica la paginación a una consulta GORM genérica.
func Paginate(db *gorm.DB, model interface{}, page, limit int, out interface{}) (*Pagination, error) {
	var totalRows int64
	pagination := &Pagination{
		Page:  page,
		Limit: limit,
	}

	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit <= 0 {
		pagination.Limit = 10 // Valor por defecto
	}

	offset := (pagination.Page - 1) * pagination.Limit

	// Contar total de registros
	if err := db.Model(model).Count(&totalRows).Error; err != nil {
		return nil, err
	}
	pagination.TotalRows = totalRows
	pagination.TotalPages = int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))

	// Obtener registros paginados
	if err := db.Model(model).Limit(pagination.Limit).Offset(offset).Find(out).Error; err != nil {
		return nil, err
	}

	pagination.Data = out
	return pagination, nil
}
