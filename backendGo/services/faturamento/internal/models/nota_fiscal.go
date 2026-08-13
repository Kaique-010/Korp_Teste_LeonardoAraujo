package models

import (
	"time"

	"gorm.io/gorm"
)

type NotaFiscal struct {
	ID           uint64           `gorm:"primaryKey" json:"id"`
	Numero       uint64           `gorm:"not null;uniqueIndex" json:"numero"`
	Status       string           `gorm:"size:10;not null" json:"status"`
	CriadoEm     time.Time        `json:"criado_em"`
	AtualizadoEm time.Time        `json:"atualizado_em"`
	FechadoEm    *time.Time       `json:"fechado_em"`
	ClienteID    *uint64          `gorm:"index" json:"cliente_id,omitempty"`
	Cliente      *Cliente         `gorm:"foreignKey:ClienteID" json:"cliente,omitempty"`
	Itens        []NotaFiscalItem `gorm:"foreignKey:NotaFiscalID" json:"itens,omitempty"`
}

func (NotaFiscal) TableName() string {
	return "nota_fiscais"
}

func (n *NotaFiscal) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	n.CriadoEm = now
	n.AtualizadoEm = now
	return nil
}

func (n *NotaFiscal) BeforeUpdate(tx *gorm.DB) error {
	n.AtualizadoEm = time.Now()
	return nil
}
