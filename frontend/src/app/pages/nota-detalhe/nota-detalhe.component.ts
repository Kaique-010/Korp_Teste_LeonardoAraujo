import { Component, OnInit, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { DatePipe, DecimalPipe, LowerCasePipe } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatDividerModule } from '@angular/material/divider';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatChipsModule } from '@angular/material/chips';

import { NotaFiscalService } from '../../services/nota-fiscal.service';
import { NotaFiscal } from '../../models/nota-fiscal.model';
import { NotaFiscalEvento } from '../../models/nota-fiscal-evento.model';
import { ItemFormDialogComponent } from '../../components/item-form-dialog/item-form-dialog.component';
import { ConfirmDialogComponent } from '../../components/confirm-dialog/confirm-dialog.component';
import { extrairMensagemErro } from '../../utils/erros';

@Component({
  selector: 'app-nota-detalhe',
  standalone: true,
  imports: [
    RouterLink,
    DatePipe,
    DecimalPipe,
    LowerCasePipe,
    MatCardModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
    MatProgressSpinnerModule,
    MatDividerModule,
    MatChipsModule,
  ],
  templateUrl: './nota-detalhe.component.html',
  styleUrls: ['./nota-detalhe.component.scss'],
})
export class NotaDetalheComponent implements OnInit {
  private readonly notaService = inject(NotaFiscalService);
  private readonly route = inject(ActivatedRoute);
  private readonly dialog = inject(MatDialog);
  private readonly snackbar = inject(MatSnackBar);

  readonly colunasItens = [
    'codigo_produto',
    'descricao_produto',
    'quantidade',
    'valor_unitario',
    'preco_vista',
    'preco_prazo',
    'desconto',
    'total',
    'acoes',
  ];
  readonly nota = signal<NotaFiscal | null>(null);
  readonly eventos = signal<NotaFiscalEvento[]>([]);
  readonly carregando = signal(false);
  readonly imprimindo = signal(false);

  private notaId = 0;

  get podeEditar(): boolean {
    return this.nota()?.status === 'ABERTA';
  }

  get totalItens(): number {
    return (this.nota()?.itens ?? []).reduce((soma, item) => soma + item.total, 0);
  }

  ngOnInit(): void {
    this.notaId = Number(this.route.snapshot.paramMap.get('id'));
    this.carregar();
  }

  carregar(): void {
    this.carregando.set(true);
    this.notaService.obter(this.notaId).subscribe({
      next: (dados) => this.nota.set(dados),
      error: (erro) => this.snackbar.open(extrairMensagemErro(erro), 'Fechar'),
      complete: () => this.carregando.set(false),
    });
    this.notaService.listarEventos(this.notaId).subscribe({
      next: (dados) => this.eventos.set(dados),
      error: (erro) => this.snackbar.open(extrairMensagemErro(erro), 'Fechar'),
    });
  }

  adicionarItem(): void {
    this.dialog
      .open(ItemFormDialogComponent, { width: '460px', data: this.notaId })
      .afterClosed()
      .subscribe((salvo) => {
        if (salvo) {
          this.carregar();
          this.snackbar.open('Item adicionado', 'Fechar');
        }
      });
  }

  removerItem(itemId: number, descricao: string): void {
    this.dialog
      .open(ConfirmDialogComponent, {
        data: {
          titulo: 'Remover item',
          mensagem: `Remover "${descricao}" da nota?`,
        },
      })
      .afterClosed()
      .subscribe((confirmado) => {
        if (!confirmado) {
          return;
        }
        this.notaService.removerItem(this.notaId, itemId).subscribe({
          next: () => {
            this.carregar();
            this.snackbar.open('Item removido', 'Fechar');
          },
          error: (erro) => this.snackbar.open(extrairMensagemErro(erro), 'Fechar'),
        });
      });
  }

  imprimir(): void {
    this.imprimindo.set(true);
    this.notaService.imprimir(this.notaId).subscribe({
      next: () => {
        this.imprimindo.set(false);
        this.snackbar.open('Impressão solicitada. A nota será fechada pelo faturamento.', 'OK');
        this.carregar();
      },
      error: (erro) => {
        this.imprimindo.set(false);
        this.snackbar.open(extrairMensagemErro(erro), 'Fechar');
      },
    });
  }
}
