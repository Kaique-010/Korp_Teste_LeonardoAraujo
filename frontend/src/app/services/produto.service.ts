import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { Produto, ProdutoInput } from '../models/produto.model';

@Injectable({ providedIn: 'root' })
export class ProdutoService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.estoqueApi}/produtos`;

  listar(): Observable<Produto[]> {
    return this.http.get<Produto[]>(this.baseUrl);
  }

  obter(id: number): Observable<Produto> {
    return this.http.get<Produto>(`${this.baseUrl}/${id}`);
  }

  criar(produto: ProdutoInput): Observable<Produto> {
    return this.http.post<Produto>(this.baseUrl, this.limparVazios(produto));
  }

  atualizar(id: number, produto: ProdutoInput): Observable<Produto> {
    return this.http.put<Produto>(`${this.baseUrl}/${id}`, this.limparVazios(produto));
  }

  excluir(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }

  private limparVazios(input: ProdutoInput): Partial<ProdutoInput> {
    const saida: Partial<ProdutoInput> = { ...input };
    if (!saida.codigo) delete saida.codigo;
    if (saida.preco_vista === null || saida.preco_vista === undefined) delete saida.preco_vista;
    if (saida.preco_prazo === null || saida.preco_prazo === undefined) delete saida.preco_prazo;
    if (saida.vigente_em === null || saida.vigente_em === undefined) delete saida.vigente_em;
    return saida;
  }
}
