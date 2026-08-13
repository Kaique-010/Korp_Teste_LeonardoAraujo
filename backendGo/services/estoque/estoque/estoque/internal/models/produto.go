package models

import (
	"time"

	"gorm.io/gorm"
)

type Produto struct {
	ID           uint64            `gorm:"primaryKey" json:"id"`
	Codigo       string            `gorm:"size:50;not null;uniqueIndex" json:"codigo"`
	Descricao    string            `gorm:"size:255;not null" json:"descricao"`
	Saldo        float64           `gorm:"not null;default:0" json:"saldo"`
	CriadoEm     time.Time         `json:"criado_em"`
	AtualizadoEm time.Time         `json:"atualizado_em"`
	Precos       []PrecoProduto    `gorm:"foreignKey:ProdutoID;constraint:OnDelete:CASCADE" json:"precos_historico,omitempty"`
	PrecoAtual   *PrecoProdutoView `gorm:"-" json:"preco_atual,omitempty"`
}

type PrecoProdutoView struct {
	PrecoVista float64 `json:"preco_vista"`
	PrecoPrazo float64 `json:"preco_prazo"`
	VigenteEm  string  `json:"vigente_em"`
}

func (p *Produto) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.CriadoEm = now
	p.AtualizadoEm = now
	return nil
}

func (p *Produto) BeforeUpdate(tx *gorm.DB) error {
	p.AtualizadoEm = time.Now()
	return nil
}
