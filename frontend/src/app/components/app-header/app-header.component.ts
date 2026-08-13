import { Component, HostListener, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ThemeToggleComponent } from '../theme-toggle/theme-toggle.component';

type NavItem = { rota: string; icone: string; rotulo: string; exact?: boolean };

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [
    RouterLink,
    RouterLinkActive,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatMenuModule,
    MatTooltipModule,
    ThemeToggleComponent,
  ],
  templateUrl: './app-header.component.html',
  styleUrls: ['./app-header.component.scss'],
})
export class AppHeaderComponent {
  readonly itens: NavItem[] = [
    { rota: '/', icone: 'dashboard', rotulo: 'Painel', exact: true },
    { rota: '/produtos', icone: 'inventory_2', rotulo: 'Produtos' },
    { rota: '/clientes', icone: 'people', rotulo: 'Clientes' },
    { rota: '/notas', icone: 'receipt_long', rotulo: 'Notas' },
  ];

  readonly mobile = signal<boolean>(this.isMobile());

  @HostListener('window:resize')
  onResize() {
    this.mobile.set(this.isMobile());
  }

  private isMobile(): boolean {
    return typeof window !== 'undefined' ? window.innerWidth < 820 : false;
  }
}
