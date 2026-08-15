import { Component, OnInit, inject, signal } from '@angular/core'
import { MatTableModule } from '@angular/material/table'
import { MatButtonModule } from '@angular/material/button'
import { MatIconModule } from '@angular/material/icon'
import { MatProgressBarModule } from '@angular/material/progress-bar'
import { MatDialog } from '@angular/material/dialog'
import { MatSnackBar } from '@angular/material/snack-bar'
import { DatePipe, DecimalPipe } from '@angular/common'

import { ProdutoService } from '../../services/produto.service'
import { Produto } from '../../models/produto.model'
import { ProdutoFormDialogComponent } from '../../components/produto-form-dialog/produto-form-dialog.component'
import { ConfirmDialogComponent } from '../../components/confirm-dialog/confirm-dialog.component'
import { extrairMensagemErro } from '../../utils/erros'

@Component({
  selector: 'app-produtos',
  standalone: true,
  imports: [
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
    DecimalPipe,
    DatePipe,
  ],
  templateUrl: './produtos.component.html',
  styleUrl: './produtos.component.scss',
})
export class ProdutosComponent implements OnInit {
  private readonly produtoService = inject(ProdutoService)
  private readonly dialog = inject(MatDialog)
  private readonly snackbar = inject(MatSnackBar)

  readonly colunas = [
    'codigo',
    'descricao',
    'saldo',
    'preco_vista',
    'preco_prazo',
    'atualizado_em',
    'acoes',
  ]
  readonly produtos = signal<Produto[]>([])
  readonly carregando = signal(false)

  ngOnInit(): void {
    this.carregar()
  }

  carregar(): void {
    this.carregando.set(true)
    this.produtoService.listar().subscribe({
      next: (dados) => this.produtos.set(dados),
      error: (erro) => this.snackbar.open(extrairMensagemErro(erro), 'Fechar'),
      complete: () => this.carregando.set(false),
    })
  }

  abrirCriar(): void {
    this.dialog
      .open(ProdutoFormDialogComponent, { width: '520px' })
      .afterClosed()
      .subscribe((salvo) => {
        if (salvo) {
          this.carregar()
          this.snackbar.open('Produto criado', 'Fechar')
        }
      })
  }

  abrirEditar(produto: Produto): void {
    this.dialog
      .open(ProdutoFormDialogComponent, { width: '520px', data: produto })
      .afterClosed()
      .subscribe((salvo) => {
        if (salvo) {
          this.carregar()
          this.snackbar.open('Produto atualizado', 'Fechar')
        }
      })
  }

  excluir(produto: Produto): void {
    this.dialog
      .open(ConfirmDialogComponent, {
        data: {
          titulo: 'Excluir produto',
          mensagem: `Excluir "${produto.descricao}" (${produto.codigo})?\nProdutos com movimentações de estoque não podem ser excluídos.`,
        },
      })
      .afterClosed()
      .subscribe((confirmado) => {
        if (!confirmado) {
          return
        }
        this.produtoService.excluir(produto.id).subscribe({
          next: () => {
            this.carregar()
            this.snackbar.open('Produto excluído', 'Fechar')
          },
          error: (erro) =>
            this.snackbar.open(extrairMensagemErro(erro), 'Fechar'),
        })
      })
  }
}
