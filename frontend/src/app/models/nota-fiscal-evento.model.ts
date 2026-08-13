export interface NotaFiscalEvento {
  id: number;
  nota_fiscal_id: number;
  tipo: string;
  descricao: string;
  dados?: Record<string, unknown>;
  criado_em: string;
}
