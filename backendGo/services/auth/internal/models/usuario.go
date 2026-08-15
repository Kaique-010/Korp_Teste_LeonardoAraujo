package models

import "time"

type Usuario struct {
	ID           uint   `gorm:"primaryKey"`
	Nome         string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	SenhaHash    string `gorm:"column:senha_hash;not null"`
	Ativo        bool   `gorm:"not null;default:true"`
	CriadoEm     time.Time
	AtualizadoEm time.Time
}
