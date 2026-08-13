package models

import (
	"time"

	"gorm.io/gorm"
)

type Cliente struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Nome      string    `gorm:"size:255;not null" json:"nome"`
	CriadoEm  time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

func (Cliente) TableName() string {
	return "clientes"
}

func (c *Cliente) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	c.CriadoEm = now
	c.AtualizadoEm = now
	return nil
}

func (c *Cliente) BeforeUpdate(tx *gorm.DB) error {
	c.AtualizadoEm = time.Now()
	return nil
}
