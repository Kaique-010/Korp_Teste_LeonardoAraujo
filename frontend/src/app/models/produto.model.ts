export interface PrecoProdutoView {
  preco_vista: number;
  preco_prazo: number;
  vigente_em?: string;
}

export interface Produto {
  id: number;
  codigo: string;
  descricao: string;
  saldo: number;
  criado_em?: string;
  atualizado_em?: string;
  preco_atual?: PrecoProdutoView | null;
}

export interface ProdutoInput {
  codigo?: string;
  descricao: string;
  saldo: number;
  preco_vista?: number | null;
  preco_prazo?: number | null;
  vigente_em?: string | null;
}
