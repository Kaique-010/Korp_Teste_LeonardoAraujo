import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, RouterLink, MatCardModule, MatButtonModule, MatIconModule],
  template: `
    <div class="home">
      <div class="intro">
        <h1>Painel Korp</h1>
        <p class="subtitulo">Sistema de notas fiscais e estoque integrado.</p>
      </div>

      <div class="cards">
        <mat-card class="card card-produtos" appearance="outlined" routerLink="/produtos">
          <mat-card-header>
            <mat-card-title>Produtos</mat-card-title>
            <mat-card-subtitle>Cadastro, saldo e preços à vista / a prazo</mat-card-subtitle>
          </mat-card-header>
          <mat-card-content>
            <mat-icon color="primary">inventory_2</mat-icon>
            <ul>
              <li>Cadastre produtos com código automático</li>
              <li>Defina preços à vista e a prazo</li>
              <li>Acompanhe o saldo de estoque</li>
            </ul>
          </mat-card-content>
          <mat-card-actions align="end">
            <button mat-flat-button color="primary" routerLink="/produtos">
              Abrir produtos
              <mat-icon>arrow_forward</mat-icon>
            </button>
          </mat-card-actions>
        </mat-card>

        <mat-card class="card card-clientes" appearance="outlined" routerLink="/clientes">
          <mat-card-header>
            <mat-card-title>Clientes</mat-card-title>
            <mat-card-subtitle>Cadastro de clientes para vincular às notas</mat-card-subtitle>
          </mat-card-header>
          <mat-card-content>
            <mat-icon color="accent">people</mat-icon>
            <ul>
              <li>Cadastro simples (ID + nome)</li>
              <li>Vincule clientes na nova nota fiscal</li>
              <li>Visualize o cliente nas listagens</li>
            </ul>
          </mat-card-content>
          <mat-card-actions align="end">
            <button mat-flat-button color="accent" routerLink="/clientes">
              Abrir clientes
              <mat-icon>arrow_forward</mat-icon>
            </button>
          </mat-card-actions>
        </mat-card>

        <mat-card class="card card-notas" appearance="outlined" routerLink="/notas">
          <mat-card-header>
            <mat-card-title>Notas fiscais</mat-card-title>
            <mat-card-subtitle>Crie notas, adicione itens e imprima (baixa estoque)</mat-card-subtitle>
          </mat-card-header>
          <mat-card-content>
            <mat-icon color="warn">receipt_long</mat-icon>
            <ul>
              <li>Nova nota com cliente (dropdown)</li>
              <li>Adicione itens com preços do produto</li>
              <li>Imprima → baixa assíncrona no estoque via RabbitMQ</li>
            </ul>
          </mat-card-content>
          <mat-card-actions align="end">
            <button mat-flat-button color="warn" routerLink="/notas">
              Abrir notas
              <mat-icon>arrow_forward</mat-icon>
            </button>
          </mat-card-actions>
        </mat-card>
      </div>
    </div>
  `,
  styles: `
    .home { max-width: 1080px; margin: 0 auto; padding: 16px; }
    .intro { text-align: center; margin-bottom: 24px; }
    .intro h1 { margin: 0; font-size: 2rem; }
    .intro .subtitulo { color: rgba(0, 0, 0, 0.6); margin-top: 4px; }
    .cards {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
      gap: 16px;
    }
    .card { cursor: pointer; transition: transform 0.15s, box-shadow 0.15s; }
    .card:hover { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(0, 0, 0, 0.12); }
    .card mat-card-content { display: flex; align-items: flex-start; gap: 16px; padding-top: 12px; }
    .card mat-card-content mat-icon {
      width: 48px; height: 48px; font-size: 40px;
      display: inline-flex; align-items: center; justify-content: center;
    }
    .card ul { margin: 0; padding-left: 20px; line-height: 1.8; }
  `,
})
export class HomeComponent {}
