package models

type NotaFiscalItem struct {
	ID               uint64  `gorm:"primaryKey" json:"id"`
	NotaFiscalID     uint64  `gorm:"not null;index" json:"nota_fiscal_id"`
	ProdutoID        uint64  `gorm:"not null" json:"produto_id"`
	CodigoProduto    string  `gorm:"size:50;not null" json:"codigo_produto"`
	DescricaoProduto string  `gorm:"size:255;not null" json:"descricao_produto"`
	Quantidade       float64 `gorm:"not null" json:"quantidade"`
	ValorUnitario    float64 `gorm:"not null" json:"valor_unitario"`
	PrecoVista       float64 `gorm:"not null;default:0" json:"preco_vista"`
	PrecoPrazo       float64 `gorm:"not null;default:0" json:"preco_prazo"`
	Desconto         float64 `gorm:"not null;default:0" json:"desconto"`
	Total            float64 `gorm:"not null" json:"total"`
}

func (NotaFiscalItem) TableName() string {
	return "nota_fiscal_itens"
}
