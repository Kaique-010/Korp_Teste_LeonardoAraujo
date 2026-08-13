package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NotaFiscalEvento struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	NotaFiscalID uint64         `gorm:"not null;index" json:"nota_fiscal_id"`
	Tipo         string         `gorm:"size:50;not null" json:"tipo"`
	Descricao    string         `gorm:"type:text;not null" json:"descricao"`
	Referencia   string         `gorm:"size:100" json:"referencia,omitempty"`
	Dados        datatypes.JSON `gorm:"type:jsonb" json:"dados,omitempty"`
	CriadoEm     time.Time      `json:"criado_em"`
}

func (NotaFiscalEvento) TableName() string {
	return "nota_fiscal_eventos"
}

func (e *NotaFiscalEvento) BeforeCreate(tx *gorm.DB) error {
	e.CriadoEm = time.Now()
	return nil
}
