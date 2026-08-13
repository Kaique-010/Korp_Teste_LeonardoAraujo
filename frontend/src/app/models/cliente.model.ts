export interface Cliente {
  id: number;
  nome: string;
  criado_em?: string;
  atualizado_em?: string;
}

export interface ClienteInput {
  nome: string;
}
