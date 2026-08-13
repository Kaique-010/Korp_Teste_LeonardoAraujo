import { Component, OnInit, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatDialog, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { CommonModule } from '@angular/common';
import { take, tap } from 'rxjs';

import { ClienteService } from '../../services/cliente.service';
import { NotaFiscalService } from '../../services/nota-fiscal.service';
import { Cliente } from '../../models/cliente.model';
import { extrairMensagemErro } from '../../utils/erros';
import { ClienteFormDialogComponent } from '../cliente-form-dialog/cliente-form-dialog.component';

@Component({
  selector: 'app-nova-nota-dialog',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatSelectModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    MatIconModule,
    MatTooltipModule,
  ],
  template: `
    <h2 mat-dialog-title>Nova nota fiscal</h2>
    <mat-dialog-content>
      <form [formGroup]="form" (ngSubmit)="confirmar()">
        @if (carregandoClientes()) {
          <div class="carregando-clientes">
            <mat-progress-spinner diameter="20" mode="indeterminate" />
            <span>Carregando clientes…</span>
          </div>
        } @else {
          <div class="campo-cliente">
            <mat-form-field class="cliente-select">
              <mat-label>Cliente (opcional)</mat-label>
              <mat-select formControlName="cliente_id">
                <mat-option [value]="null">Sem cliente</mat-option>
                @for (c of clientes(); track c.id) {
                  <mat-option [value]="c.id">{{ c.nome }}</mat-option>
                }
              </mat-select>
            </mat-form-field>
            <button
              type="button"
              mat-icon-button
              color="primary"
              (click)="abrirNovoCliente()"
              matTooltip="Cadastrar novo cliente"
            >
              <mat-icon>add</mat-icon>
            </button>
          </div>
        }
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button type="button" (click)="dialogRef.close()">Cancelar</button>
      <button mat-flat-button color="primary" [disabled]="criando()" (click)="confirmar()">
        @if (criando()) {
          <mat-progress-spinner diameter="18" mode="indeterminate" />
        }
        Criar nota
      </button>
    </mat-dialog-actions>
  `,
  styles: `
    .carregando-clientes {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 0;
    }
    .campo-cliente {
      display: flex;
      align-items: end;
      gap: 8px;
      .cliente-select {
        flex: 1;
      }
    }
    mat-dialog-content {
      padding-top: 8px;
    }
  `,
})
export class NovaNotaDialogComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly clienteService = inject(ClienteService);
  private readonly notaService = inject(NotaFiscalService);
  readonly dialogRef = inject(MatDialogRef<NovaNotaDialogComponent>);
  private readonly dialog = inject(MatDialog);
  private readonly snackbar = inject(MatSnackBar);
  private readonly router = inject(Router);

  readonly clientes = signal<Cliente[]>([]);
  readonly carregandoClientes = signal(true);
  readonly criando = signal(false);
  readonly form = this.fb.group({
    cliente_id: <number | null>null,
  });

  ngOnInit(): void {
    this.carregarClientes();
  }

  private carregarClientes(): void {
    this.clienteService
      .listar()
      .pipe(take(1))
      .subscribe({
        next: (dados) => {
          this.clientes.set(dados);
          this.carregandoClientes.set(false);
        },
        error: (erro) => {
          this.snackbar.open(extrairMensagemErro(erro), 'Fechar');
          this.carregandoClientes.set(false);
        },
      });
  }

  abrirNovoCliente(): void {
    this.dialog
      .open(ClienteFormDialogComponent, { width: '420px' })
      .afterClosed()
      .pipe(take(1))
      .subscribe((salvo) => {
        if (salvo) {
          this.carregarClientes();
          this.snackbar.open('Cliente criado. Selecione-o na lista.', 'Fechar');
        }
      });
  }

  confirmar(): void {
    this.criando.set(true);
    const clienteId = this.form.value.cliente_id ?? null;
    this.notaService
      .criar({ cliente_id: clienteId })
      .pipe(
        tap({
          next: (nota) => {
            this.dialogRef.close(true);
            this.router.navigate(['/notas', nota.id]);
          },
          error: (erro) => {
            this.criando.set(false);
            this.snackbar.open(extrairMensagemErro(erro), 'Fechar');
          },
        }),
        take(1),
      )
      .subscribe();
  }
}
