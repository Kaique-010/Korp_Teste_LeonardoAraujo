import { HttpErrorResponse } from '@angular/common/http';

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

/**
 * Extrai a mensagem amigável de um erro. O backend sempre responde
 * `{ "error": { code, message, details } }`, mas também tratamos erros
 * de rede (sem resposta) e erros locais.
 */
export function extrairMensagemErro(erro: unknown): string {
  if (erro instanceof HttpErrorResponse) {
    const corpo = erro.error as { error?: ApiError } | null;
    if (corpo?.error?.message) {
      return corpo.error.message;
    }
    if (erro.status === 0) {
      return 'Não foi possível conectar ao serviço. Verifique se ele está no ar.';
    }
    if (erro.status) {
      return `Erro ${erro.status} do servidor`;
    }
  }
  if (erro instanceof Error) {
    return erro.message;
  }
  return 'Erro inesperado';
}
