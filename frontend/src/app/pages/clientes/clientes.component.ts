import { Component, OnInit, inject, signal } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { DatePipe } from '@angular/common';

import { ClienteService } from '../../services/cliente.service';
import { Cliente } from '../../models/cliente.model';
import { ClienteFormDialogComponent } from '../../components/cliente-form-dialog/cliente-form-dialog.component';
import { ConfirmDialogComponent } from '../../components/confirm-dialog/confirm-dialog.component';
import { extrairMensagemErro } from '../../utils/erros';

@Component({
  selector: 'app-clientes',
  standalone: true,
  imports: [
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
    DatePipe,
  ],
  templateUrl: './clientes.component.html',
  styleUrls: ['./clientes.component.scss'],
})
export class ClientesComponent implements OnInit {
  private readonly clienteService = inject(ClienteService);
  private readonly dialog = inject(MatDialog);
  private readonly snackbar = inject(MatSnackBar);

  readonly colunas = ['id', 'nome', 'criado_em', 'acoes'];
  readonly clientes = signal<Cliente[]>([]);
  readonly carregando = signal(false);

  ngOnInit(): void {
    this.carregar();
  }

  carregar(): void {
    this.carregando.set(true);
    this.clienteService.listar().subscribe({
      next: (dados) => this.clientes.set(dados),
      error: (erro) => this.snackbar.open(extrairMensagemErro(erro), 'Fechar'),
      complete: () => this.carregando.set(false),
    });
  }

  abrirCriar(): void {
    this.dialog
      .open(ClienteFormDialogComponent, { width: '420px' })
      .afterClosed()
      .subscribe((salvo) => {
        if (salvo) {
          this.carregar();
          this.snackbar.open('Cliente criado', 'Fechar');
        }
      });
  }

  abrirEditar(cliente: Cliente): void {
    this.dialog
      .open(ClienteFormDialogComponent, { width: '420px', data: cliente })
      .afterClosed()
      .subscribe((salvo) => {
        if (salvo) {
          this.carregar();
          this.snackbar.open('Cliente atualizado', 'Fechar');
        }
      });
  }

  excluir(cliente: Cliente): void {
    this.dialog
      .open(ConfirmDialogComponent, {
        data: {
          titulo: 'Excluir cliente',
          mensagem: `Excluir "${cliente.nome}"?\nClientes com notas fiscais não podem ser excluídos.`,
        },
      })
      .afterClosed()
      .subscribe((confirmado) => {
        if (!confirmado) return;
        this.clienteService.excluir(cliente.id).subscribe({
          next: () => {
            this.carregar();
            this.snackbar.open('Cliente excluído', 'Fechar');
          },
          error: (erro) => this.snackbar.open(extrairMensagemErro(erro), 'Fechar'),
        });
      });
  }
}
