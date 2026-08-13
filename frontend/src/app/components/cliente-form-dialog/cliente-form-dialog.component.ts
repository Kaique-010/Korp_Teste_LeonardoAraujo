import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';

import { ClienteService } from '../../services/cliente.service';
import { Cliente, ClienteInput } from '../../models/cliente.model';
import { extrairMensagemErro } from '../../utils/erros';

@Component({
  selector: 'app-cliente-form-dialog',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressSpinnerModule,
  ],
  template: `
    <h2 mat-dialog-title>{{ cliente ? 'Editar cliente' : 'Novo cliente' }}</h2>
    <mat-dialog-content>
      <form [formGroup]="formulario" (ngSubmit)="salvar()">
        <mat-form-field>
          <mat-label>Nome</mat-label>
          <input matInput formControlName="nome" placeholder="Nome completo" />
          @if (nome.hasError('required')) {
            <mat-error>Nome é obrigatório</mat-error>
          }
          @if (nome.hasError('maxlength')) {
            <mat-error>Máximo 255 caracteres</mat-error>
          }
        </mat-form-field>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button type="button" (click)="fechar()">Cancelar</button>
      <button
        mat-flat-button
        color="primary"
        [disabled]="formulario.invalid || salvando()"
        (click)="salvar()"
      >
        @if (salvando()) {
          <mat-progress-spinner diameter="18" mode="indeterminate" />
        }
        Salvar
      </button>
    </mat-dialog-actions>
  `,
  styles: `
    mat-dialog-content form {
      display: flex;
      flex-direction: column;
      gap: 4px;
      padding-top: 8px;
    }
  `,
})
export class ClienteFormDialogComponent {
  private readonly fb = inject(FormBuilder);
  private readonly clienteService = inject(ClienteService);
  private readonly dialogRef = inject(MatDialogRef<ClienteFormDialogComponent>);
  private readonly snackbar = inject(MatSnackBar);
  protected readonly cliente = inject<Cliente | undefined>(MAT_DIALOG_DATA);

  readonly salvando = signal(false);

  readonly formulario = this.fb.group({
    nome: [this.cliente?.nome ?? '', [Validators.required, Validators.maxLength(255)]],
  });

  get nome() {
    return this.formulario.controls.nome;
  }

  salvar(): void {
    if (this.formulario.invalid) return;
    this.salvando.set(true);
    const valores = this.formulario.value as ClienteInput;
    const obs$ = this.cliente
      ? this.clienteService.atualizar(this.cliente.id, valores)
      : this.clienteService.criar(valores);

    obs$.subscribe({
      next: () => this.fechar(true),
      error: (erro) => {
        this.salvando.set(false);
        this.snackbar.open(extrairMensagemErro(erro), 'Fechar');
      },
    });
  }

  fechar(salvo = false): void {
    this.dialogRef.close(salvo);
  }
}
