package models

import (
	"time"

	"gorm.io/gorm"
)

type MovimentoEstoque struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	ProdutoID      uint64    `gorm:"not null;index" json:"produto_id"`
	Tipo           string    `gorm:"size:10;not null" json:"tipo"`
	Quantidade     float64   `gorm:"not null" json:"quantidade"`
	Origem         string    `gorm:"size:50;not null" json:"origem"`
	Referencia     string    `gorm:"size:100" json:"referencia"`
	IdempotencyKey string    `gorm:"size:100;index" json:"idempotency_key"`
	CriadoEm       time.Time `json:"criado_em"`
}

func (MovimentoEstoque) TableName() string {
	return "movimentos_estoque"
}

func (m *MovimentoEstoque) BeforeCreate(tx *gorm.DB) error {
	m.CriadoEm = time.Now()
	return nil
}
