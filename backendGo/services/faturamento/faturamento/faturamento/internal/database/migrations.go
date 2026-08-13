package database

import "gorm.io/gorm"

const notaFiscalSequenceSQL = `
CREATE SEQUENCE IF NOT EXISTS nota_fiscal_numero_seq;
`

func ApplyNotaFiscalSequence(db *gorm.DB) error {
	return db.Exec(notaFiscalSequenceSQL).Error
}
