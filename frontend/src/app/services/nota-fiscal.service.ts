import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { CriarNotaInput, NotaFiscal } from '../models/nota-fiscal.model';
import { ItemInput } from '../models/nota-fiscal-item.model';
import { NotaFiscalEvento } from '../models/nota-fiscal-evento.model';

@Injectable({ providedIn: 'root' })
export class NotaFiscalService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.faturamentoApi}/notas`;

  listar(): Observable<NotaFiscal[]> {
    return this.http.get<NotaFiscal[]>(this.baseUrl);
  }

  obter(id: number): Observable<NotaFiscal> {
    return this.http.get<NotaFiscal>(`${this.baseUrl}/${id}`);
  }

  criar(input?: CriarNotaInput): Observable<NotaFiscal> {
    const corpo = input && input.cliente_id ? { cliente_id: input.cliente_id } : {};
    return this.http.post<NotaFiscal>(this.baseUrl, corpo);
  }

  adicionarItem(notaId: number, item: ItemInput): Observable<NotaFiscal> {
    return this.http.post<NotaFiscal>(`${this.baseUrl}/${notaId}/itens`, item);
  }

  removerItem(notaId: number, itemId: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${notaId}/itens/${itemId}`);
  }

  imprimir(notaId: number): Observable<NotaFiscal> {
    return this.http.post<NotaFiscal>(`${this.baseUrl}/${notaId}/imprimir`, {});
  }

  listarEventos(notaId: number): Observable<NotaFiscalEvento[]> {
    return this.http.get<NotaFiscalEvento[]>(`${this.baseUrl}/${notaId}/eventos`);
  }
}
