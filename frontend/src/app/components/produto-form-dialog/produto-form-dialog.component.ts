import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatSnackBar } from '@angular/material/snack-bar';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { ProdutoService } from '../../services/produto.service';
import { Produto, ProdutoInput } from '../../models/produto.model';
import { extrairMensagemErro } from '../../utils/erros';

type TipoPreco = 'vista' | 'prazo' | 'ambos';

@Component({
  selector: 'app-produto-form-dialog',
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    MatButtonToggleModule,
  ],
  template: `
    <h2 mat-dialog-title>{{ produto ? 'Editar produto' : 'Novo produto' }}</h2>

    <mat-dialog-content>
      <form [formGroup]="formulario" (ngSubmit)="salvar()">
        <mat-form-field>
          <mat-label>Código</mat-label>
          <input matInput formControlName="codigo" placeholder="PROD-000001 (auto se vazio)" />
          @if (codigo.hasError('maxlength')) {
            <mat-error>Máximo 50 caracteres</mat-error>
          }
        </mat-form-field>

        <mat-form-field>
          <mat-label>Descrição</mat-label>
          <input matInput formControlName="descricao" placeholder="Produto de exemplo" />
          @if (descricao.hasError('required')) {
            <mat-error>Descrição é obrigatória</mat-error>
          }
        </mat-form-field>

        <mat-form-field>
          <mat-label>Saldo inicial</mat-label>
          <input matInput type="number" formControlName="saldo" />
          @if (saldo.hasError('min')) {
            <mat-error>Saldo não pode ser negativo</mat-error>
          }
        </mat-form-field>

        <div class="secao-preco">
          <div class="tipo-preco-linha">
            <span class="rotulo-tipo">Tipo de preço aplicado:</span>
            <mat-button-toggle-group formControlName="tipoPreco" aria-label="Tipo de preço">
              <mat-button-toggle value="vista">A vista</mat-button-toggle>
              <mat-button-toggle value="prazo">A prazo</mat-button-toggle>
              <mat-button-toggle value="ambos">Ambos</mat-button-toggle>
            </mat-button-toggle-group>
          </div>

          <div class="linha-precos">
            <mat-form-field>
              <mat-label>Preço à vista (R$)</mat-label>
              <input
                matInput
                type="number"
                step="0.01"
                formControlName="preco_vista"
                min="0"
                [disabled]="tipoPreco.value === 'prazo'"
              />
              @if (precoVista.hasError('min')) {
                <mat-error>Não pode ser negativo</mat-error>
              }
              @if (precoVista.hasError('required')) {
                <mat-error>Informe o preço à vista</mat-error>
              }
            </mat-form-field>

            <mat-form-field>
              <mat-label>Preço a prazo (R$)</mat-label>
              <input
                matInput
                type="number"
                step="0.01"
                formControlName="preco_prazo"
                min="0"
                [disabled]="tipoPreco.value === 'vista'"
              />
              @if (precoPrazo.hasError('min')) {
                <mat-error>Não pode ser negativo</mat-error>
              }
              @if (precoPrazo.hasError('required')) {
                <mat-error>Informe o preço a prazo</mat-error>
              }
            </mat-form-field>
          </div>
        </div>
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
    .secao-preco {
      margin-top: 8px;
      display: flex;
      flex-direction: column;
      gap: 8px;
    }
    .tipo-preco-linha {
      display: flex;
      align-items: center;
      gap: 12px;
      flex-wrap: wrap;
    }
    .rotulo-tipo {
      font-size: 13px;
      color: var(--mat-sys-on-surface-variant, #666);
    }
    .linha-precos {
      display: flex;
      gap: 12px;
      mat-form-field {
        flex: 1;
      }
    }
    @media (max-width: 520px) {
      .linha-precos {
        flex-direction: column;
        gap: 0;
      }
    }
  `,
})
export class ProdutoFormDialogComponent {
  private readonly fb = inject(FormBuilder);
  private readonly produtoService = inject(ProdutoService);
  private readonly dialogRef = inject(MatDialogRef<ProdutoFormDialogComponent>);
  private readonly snackbar = inject(MatSnackBar);
  protected readonly produto = inject<Produto | undefined>(MAT_DIALOG_DATA);

  readonly salvando = signal(false);

  private readonly modoInicial: TipoPreco = this.derivarModoInicial();

  readonly formulario = this.fb.group({
    codigo: [this.produto?.codigo ?? '', [Validators.maxLength(50)]],
    descricao: [this.produto?.descricao ?? '', [Validators.required, Validators.maxLength(255)]],
    saldo: [this.produto?.saldo ?? 0, [Validators.required, Validators.min(0)]],
    tipoPreco: [this.modoInicial as TipoPreco, [Validators.required]],
    preco_vista: [
      this.produto?.preco_atual?.preco_vista ?? (null as number | null),
      [Validators.min(0)],
    ],
    preco_prazo: [
      this.produto?.preco_atual?.preco_prazo ?? (null as number | null),
      [Validators.min(0)],
    ],
  });

  get codigo() {
    return this.formulario.controls.codigo;
  }
  get descricao() {
    return this.formulario.controls.descricao;
  }
  get saldo() {
    return this.formulario.controls.saldo;
  }
  get tipoPreco() {
    return this.formulario.controls.tipoPreco;
  }
  get precoVista() {
    return this.formulario.controls.preco_vista;
  }
  get precoPrazo() {
    return this.formulario.controls.preco_prazo;
  }

  constructor() {
    this.tipoPreco.valueChanges
      .pipe(takeUntilDestroyed())
      .subscribe(() => this.aplicarValidadoresCondicionais());
    this.aplicarValidadoresCondicionais();
  }

  salvar(): void {
    if (this.formulario.invalid) {
      return;
    }
    this.salvando.set(true);

    const modo = this.tipoPreco.value as TipoPreco;
    const raw = this.formulario.getRawValue();
    const input: ProdutoInput = {
      codigo: raw.codigo?.trim() || undefined,
      descricao: raw.descricao!,
      saldo: raw.saldo ?? 0,
      preco_vista: modo === 'prazo' ? null : (raw.preco_vista ?? null),
      preco_prazo: modo === 'vista' ? null : (raw.preco_prazo ?? null),
    };

    const operacao = this.produto
      ? this.produtoService.atualizar(this.produto.id, input)
      : this.produtoService.criar(input);

    operacao.subscribe({
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

  private aplicarValidadoresCondicionais(): void {
    const modo = this.tipoPreco.value as TipoPreco;

    if (modo === 'vista' || modo === 'ambos') {
      this.precoVista.setValidators([Validators.min(0), Validators.required]);
      this.precoVista.enable({ emitEvent: false });
    } else {
      this.precoVista.setValidators([Validators.min(0)]);
      this.precoVista.setValue(null, { emitEvent: false });
      this.precoVista.disable({ emitEvent: false });
    }

    if (modo === 'prazo' || modo === 'ambos') {
      this.precoPrazo.setValidators([Validators.min(0), Validators.required]);
      this.precoPrazo.enable({ emitEvent: false });
    } else {
      this.precoPrazo.setValidators([Validators.min(0)]);
      this.precoPrazo.setValue(null, { emitEvent: false });
      this.precoPrazo.disable({ emitEvent: false });
    }

    this.precoVista.updateValueAndValidity({ emitEvent: false });
    this.precoPrazo.updateValueAndValidity({ emitEvent: false });
  }

  private derivarModoInicial(): TipoPreco {
    const p = this.produto?.preco_atual;
    if (!p) return 'ambos';
    const temVista = typeof p.preco_vista === 'number' && p.preco_vista > 0;
    const temPrazo = typeof p.preco_prazo === 'number' && p.preco_prazo > 0;
    if (temVista && temPrazo) return 'ambos';
    if (temVista) return 'vista';
    if (temPrazo) return 'prazo';
    return 'ambos';
  }
}
