package database

import "gorm.io/gorm"

const movimentoTriggerSQL = `
DROP TRIGGER IF EXISTS trg_movimento_estoque_insert ON movimentos_estoque;

CREATE OR REPLACE FUNCTION trg_movimento_estoque_fn() RETURNS TRIGGER AS $$
DECLARE
    saldo_atual NUMERIC;
BEGIN
    IF NEW.tipo NOT IN ('ENTRADA', 'SAIDA') THEN
        RAISE EXCEPTION 'TIPO_INVALIDO';
    END IF;

    IF NEW.quantidade <= 0 THEN
        RAISE EXCEPTION 'QUANTIDADE_INVALIDA';
    END IF;

    SELECT saldo INTO saldo_atual
    FROM produtos
    WHERE id = NEW.produto_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'PRODUTO_NAO_ENCONTRADO';
    END IF;

    IF NEW.tipo = 'SAIDA' AND saldo_atual < NEW.quantidade THEN
        RAISE EXCEPTION 'SALDO_INSUFICIENTE';
    END IF;

    UPDATE produtos
    SET saldo = saldo + CASE
            WHEN NEW.tipo = 'ENTRADA' THEN NEW.quantidade
            ELSE -NEW.quantidade
        END,
        atualizado_em = now()
    WHERE id = NEW.produto_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_movimento_estoque_insert
BEFORE INSERT ON movimentos_estoque
FOR EACH ROW
EXECUTE FUNCTION trg_movimento_estoque_fn();
`

func ApplyMovimentoTrigger(db *gorm.DB) error {
	return db.Exec(movimentoTriggerSQL).Error
}

// Constraint de idempotência: cada chave só pode gerar um movimento.
// Índice parcial ignora chaves vazias (movimentos manuais/HTTP sem chave).
func ApplyMovimentoConstraints(db *gorm.DB) error {
	return db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS uq_movimentos_idempotency
ON movimentos_estoque (idempotency_key)
WHERE idempotency_key <> '';
`).Error
}
