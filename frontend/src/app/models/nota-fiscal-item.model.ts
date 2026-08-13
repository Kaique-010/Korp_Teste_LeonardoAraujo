export interface NotaFiscalItem {
  id: number;
  nota_fiscal_id: number;
  produto_id: number;
  codigo_produto: string;
  descricao_produto: string;
  quantidade: number;
  valor_unitario: number;
  desconto: number;
  total: number;
  preco_vista?: number | null;
  preco_prazo?: number | null;
}

export interface ItemInput {
  produto_id: number;
  quantidade: number;
  valor_unitario: number;
  desconto: number;
  preco_vista?: number | null;
  preco_prazo?: number | null;
}
