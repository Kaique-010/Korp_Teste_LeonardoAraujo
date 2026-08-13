package models

import (
	"time"

	"gorm.io/gorm"
)

// PrecoProduto registra o histórico de preços de um produto.
// Quando um novo preço é registrado, o anterior tem fim_em definido.
type PrecoProduto struct {
	ID         uint64     `gorm:"primaryKey" json:"id"`
	ProdutoID  uint64     `gorm:"not null;index" json:"produto_id"`
	PrecoVista float64    `gorm:"not null;default:0" json:"preco_vista"`
	PrecoPrazo float64    `gorm:"not null;default:0" json:"preco_prazo"`
	VigenteEm  time.Time  `gorm:"not null;default:now()" json:"vigente_em"`
	FimEm      *time.Time `json:"fim_em,omitempty"`
	CriadoEm   time.Time  `json:"criado_em"`
}

func (PrecoProduto) TableName() string {
	return "precos_produtos"
}

func (p *PrecoProduto) BeforeCreate(tx *gorm.DB) error {
	if p.VigenteEm.IsZero() {
		p.VigenteEm = time.Now()
	}
	p.CriadoEm = time.Now()
	return nil
}
