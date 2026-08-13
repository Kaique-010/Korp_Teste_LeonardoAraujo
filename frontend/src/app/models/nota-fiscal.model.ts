import { NotaFiscalItem } from './nota-fiscal-item.model';
import { Cliente } from './cliente.model';

export type NotaStatus = 'ABERTA' | 'FECHADA';

export interface NotaFiscal {
  id: number;
  numero: number;
  status: NotaStatus;
  criado_em: string;
  atualizado_em: string;
  fechado_em?: string | null;
  cliente_id?: number | null;
  cliente?: Cliente | null;
  itens?: NotaFiscalItem[];
}

export interface CriarNotaInput {
  cliente_id?: number | null;
}
