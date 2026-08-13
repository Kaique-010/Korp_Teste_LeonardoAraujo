import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatSnackBar } from '@angular/material/snack-bar';
import { DecimalPipe } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { ProdutoService } from '../../services/produto.service';
import { NotaFiscalService } from '../../services/nota-fiscal.service';
import { Produto } from '../../models/produto.model';
import { ItemInput } from '../../models/nota-fiscal-item.model';
import { extrairMensagemErro } from '../../utils/erros';

type ModoPreco = 'vista' | 'prazo' | 'manual';

@Component({
  selector: 'app-item-form-dialog',
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    MatButtonToggleModule,
    DecimalPipe,
  ],
  template: `
    <h2 mat-dialog-title>Adicionar item</h2>

    <mat-dialog-content>
      <form [formGroup]="formulario" (ngSubmit)="salvar()">
        <mat-form-field>
          <mat-label>Produto</mat-label>
          <mat-select formControlName="produtoId">
            @for (produto of produtos(); track produto.id) {
              <mat-option [value]="produto.id">
                {{ produto.codigo }} — {{ produto.descricao }} (saldo {{ produto.saldo }})
              </mat-option>
            }
          </mat-select>
          @if (produtoId.hasError('required')) {
            <mat-error>Selecione um produto</mat-error>
          }
        </mat-form-field>

        @if (produtoSelecionado()) {
          <div class="preco-aplicado">
            <span class="rotulo-preco">Usar preço do produto:</span>
            <mat-button-toggle-group
              formControlName="modoPreco"
              aria-label="Modo de preço do item"
              [disabled]="!temAmbosPrecos() && !temPrecoSelecionado()"
            >
              @if (produtoSelecionado()?.preco_atual?.preco_vista != null) {
                <mat-button-toggle value="vista">
                  A vista
                  <small
                    >(R$
                    {{ produtoSelecionado()!.preco_atual!.preco_vista! | number: '1.2-2' }})</small
                  >
                </mat-button-toggle>
              }
              @if (produtoSelecionado()?.preco_atual?.preco_prazo != null) {
                <mat-button-toggle value="prazo">
                  A prazo
                  <small
                    >(R$
                    {{ produtoSelecionado()!.preco_atual!.preco_prazo! | number: '1.2-2' }})</small
                  >
                </mat-button-toggle>
              }
              <mat-button-toggle value="manual"> Manual </mat-button-toggle>
            </mat-button-toggle-group>
          </div>
        }

        <div class="linha-qtd-valor">
          <mat-form-field>
            <mat-label>Quantidade</mat-label>
            <input matInput type="number" formControlName="quantidade" />
            @if (quantidade.hasError('min')) {
              <mat-error>Quantidade deve ser maior que zero</mat-error>
            }
          </mat-form-field>

          <mat-form-field>
            <mat-label>Valor unitário</mat-label>
            <input
              matInput
              type="number"
              step="0.01"
              formControlName="valorUnitario"
              [disabled]="formulario.controls.modoPreco.value !== 'manual'"
            />
            @if (valorUnitario.hasError('min')) {
              <mat-error>Valor deve ser maior que zero</mat-error>
            }
            @if (valorUnitario.hasError('required')) {
              <mat-error>Informe o valor unitário</mat-error>
            }
          </mat-form-field>
        </div>

        <mat-form-field>
          <mat-label>Desconto (%)</mat-label>
          <input matInput type="number" step="0.01" formControlName="desconto" />
          @if (desconto.hasError('min') || desconto.hasError('max')) {
            <mat-error>Desconto deve ser entre 0 e 100</mat-error>
          }
        </mat-form-field>
      </form>

      <p class="total">
        Total do item:
        <strong>{{ totalItem() | number: '1.2-2' }}</strong>
      </p>
    </mat-dialog-content>

    <mat-dialog-actions align="end">
      <button mat-button type="button" (click)="fechar()">Cancelar</button>
      <button
        mat-flat-button
        color="primary"
        [disabled]="
          formulario.invalid || salvando() || produtos().length === 0 || !valorUnitarioValido()
        "
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
    .preco-aplicado {
      margin: 8px 0 4px;
      display: flex;
      flex-direction: column;
      gap: 8px;
    }
    .rotulo-preco {
      font-size: 13px;
      color: var(--mat-sys-on-surface-variant, #666);
    }
    ::ng-deep mat-button-toggle small {
      margin-left: 6px;
      opacity: 0.85;
      font-size: 11px;
    }
    .linha-qtd-valor {
      display: flex;
      gap: 12px;
      mat-form-field {
        flex: 1;
      }
    }
    .total {
      margin-top: 12px;
    }
    @media (max-width: 520px) {
      .linha-qtd-valor {
        flex-direction: column;
        gap: 0;
      }
    }
  `,
})
export class ItemFormDialogComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly produtoService = inject(ProdutoService);
  private readonly notaService = inject(NotaFiscalService);
  private readonly dialogRef = inject(MatDialogRef<ItemFormDialogComponent>);
  private readonly snackbar = inject(MatSnackBar);
  private readonly notaId = inject<number>(MAT_DIALOG_DATA);

  readonly produtos = signal<Produto[]>([]);
  readonly salvando = signal(false);
  readonly produtoSelecionado = signal<Produto | undefined>(undefined);
  readonly totalItem = signal<number>(0);

  readonly formulario = this.fb.group({
    produtoId: [null as number | null, [Validators.required]],
    modoPreco: ['manual' as ModoPreco, [Validators.required]],
    quantidade: [1, [Validators.required, Validators.min(0.000001)]],
    valorUnitario: [0 as number | null, [Validators.min(0)]],
    desconto: [0, [Validators.min(0), Validators.max(100)]],
  });

  get produtoId() {
    return this.formulario.controls.produtoId;
  }
  get modoPreco() {
    return this.formulario.controls.modoPreco;
  }
  get quantidade() {
    return this.formulario.controls.quantidade;
  }
  get valorUnitario() {
    return this.formulario.controls.valorUnitario;
  }
  get desconto() {
    return this.formulario.controls.desconto;
  }

  constructor() {
    this.produtoId.valueChanges.pipe(takeUntilDestroyed()).subscribe((id) => {
      const sel = this.produtos().find((p) => p.id === id);
      this.produtoSelecionado.set(sel);
      this.ajustarModoAoSelecionarProduto(sel);
    });

    this.modoPreco.valueChanges
      .pipe(takeUntilDestroyed())
      .subscribe(() => this.aplicarPrecoPeloModo());

    this.formulario.valueChanges.pipe(takeUntilDestroyed()).subscribe(() => {
      const q = this.quantidade.value ?? 0;
      const v = this.valorUnitario.value ?? 0;
      const d = this.desconto.value ?? 0;
      this.totalItem.set(q * (v ?? 0) * (1 - d / 100));
    });
  }

  ngOnInit(): void {
    this.produtoService.listar().subscribe({
      next: (dados) => {
        this.produtos.set(dados);
        const idAtual = this.produtoId.value;
        if (idAtual != null) {
          this.produtoSelecionado.set(dados.find((p) => p.id === idAtual));
        }
      },
      error: (erro) => this.snackbar.open(extrairMensagemErro(erro), 'Fechar'),
    });
  }

  salvar(): void {
    if (!this.valorUnitarioValido() || this.formulario.invalid) {
      return;
    }
    this.salvando.set(true);

    const modo = this.modoPreco.value as ModoPreco;
    const precoSelecionado = this.produtoSelecionado()?.preco_atual;
    const preco_vista = modo === 'vista' ? (precoSelecionado?.preco_vista ?? null) : null;
    const preco_prazo = modo === 'prazo' ? (precoSelecionado?.preco_prazo ?? null) : null;
    const valor_unitario = this.valorUnitario.value ?? 0;

    const item: ItemInput = {
      produto_id: this.produtoId.value!,
      quantidade: this.quantidade.value!,
      valor_unitario,
      desconto: this.desconto.value ?? 0,
      preco_vista,
      preco_prazo,
    };
    this.notaService.adicionarItem(this.notaId, item).subscribe({
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

  temAmbosPrecos(): boolean {
    const p = this.produtoSelecionado()?.preco_atual;
    return !!p && typeof p.preco_vista === 'number' && typeof p.preco_prazo === 'number';
  }

  temPrecoSelecionado(): boolean {
    const p = this.produtoSelecionado()?.preco_atual;
    return !!p && (typeof p.preco_vista === 'number' || typeof p.preco_prazo === 'number');
  }

  valorUnitarioValido(): boolean {
    const v = this.valorUnitario.value;
    return typeof v === 'number' && v > 0;
  }

  private ajustarModoAoSelecionarProduto(produto?: Produto): void {
    const p = produto?.preco_atual;
    if (!p) {
      this.modoPreco.setValue('manual', { emitEvent: true });
      return;
    }
    const temVista = typeof p.preco_vista === 'number';
    const temPrazo = typeof p.preco_prazo === 'number';
    if (temVista && temPrazo) {
      this.modoPreco.setValue('vista', { emitEvent: true });
    } else if (temVista) {
      this.modoPreco.setValue('vista', { emitEvent: true });
    } else if (temPrazo) {
      this.modoPreco.setValue('prazo', { emitEvent: true });
    } else {
      this.modoPreco.setValue('manual', { emitEvent: true });
    }
  }

  private aplicarPrecoPeloModo(): void {
    const modo = this.modoPreco.value as ModoPreco;
    const p = this.produtoSelecionado()?.preco_atual;

    if (modo === 'vista' && p && typeof p.preco_vista === 'number') {
      this.valorUnitario.setValue(p.preco_vista, { emitEvent: true });
      this.valorUnitario.disable({ emitEvent: false });
      return;
    }
    if (modo === 'prazo' && p && typeof p.preco_prazo === 'number') {
      this.valorUnitario.setValue(p.preco_prazo, { emitEvent: true });
      this.valorUnitario.disable({ emitEvent: false });
      return;
    }
    this.valorUnitario.enable({ emitEvent: false });
  }
}
