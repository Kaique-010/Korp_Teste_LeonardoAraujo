import { Component, OnInit, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { DatePipe, LowerCasePipe } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatDialog } from '@angular/material/dialog';

import { NotaFiscalService } from '../../services/nota-fiscal.service';
import { NotaFiscal } from '../../models/nota-fiscal.model';
import { extrairMensagemErro } from '../../utils/erros';
import { NovaNotaDialogComponent } from '../../components/nova-nota-dialog/nova-nota-dialog.component';

@Component({
  selector: 'app-notas',
  standalone: true,
  imports: [
    DatePipe,
    LowerCasePipe,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './notas.component.html',
  styleUrls: ['./notas.component.scss'],
})
export class NotasComponent implements OnInit {
  private readonly notaService = inject(NotaFiscalService);
  private readonly router = inject(Router);
  private readonly snackbar = inject(MatSnackBar);
  private readonly dialog = inject(MatDialog);

  readonly colunas = ['id', 'numero', 'cliente', 'status', 'criado_em', 'acoes'];
  readonly notas = signal<NotaFiscal[]>([]);
  readonly carregando = signal(false);

  ngOnInit(): void {
    this.carregar();
  }

  carregar(): void {
    this.carregando.set(true);
    this.notaService.listar().subscribe({
      next: (dados) => this.notas.set(dados),
      error: (erro) => this.snackbar.open(extrairMensagemErro(erro), 'Fechar'),
      complete: () => this.carregando.set(false),
    });
  }

  criarNota(): void {
    this.dialog
      .open(NovaNotaDialogComponent, { width: '480px', disableClose: true })
      .afterClosed()
      .subscribe((feito) => {
        if (feito) this.carregar();
      });
  }

  abrirDetalhe(nota: NotaFiscal): void {
    this.router.navigate(['/notas', nota.id]);
  }
}
