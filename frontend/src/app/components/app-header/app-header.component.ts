import { Component, HostListener, inject, signal } from '@angular/core'
import { Router, RouterLink, RouterLinkActive } from '@angular/router'
import { MatToolbarModule } from '@angular/material/toolbar'
import { MatButtonModule } from '@angular/material/button'
import { MatIconModule } from '@angular/material/icon'
import { MatMenuModule } from '@angular/material/menu'
import { MatDividerModule } from '@angular/material/divider'
import { MatTooltipModule } from '@angular/material/tooltip'
import { ThemeToggleComponent } from '../theme-toggle/theme-toggle.component'
import { AuthService } from '../../services/auth.service'
import { CommonModule } from '@angular/common'

type NavItem = { rota: string; icone: string; rotulo: string; exact?: boolean }

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    RouterLinkActive,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatMenuModule,
    MatDividerModule,
    MatTooltipModule,
    ThemeToggleComponent,
  ],
  templateUrl: './app-header.component.html',
  styleUrls: ['./app-header.component.scss'],
})
export class AppHeaderComponent {
  private readonly auth = inject(AuthService)
  private readonly router = inject(Router)

  readonly itens: NavItem[] = [
    { rota: '/', icone: 'dashboard', rotulo: 'Painel', exact: true },
    { rota: '/produtos', icone: 'inventory_2', rotulo: 'Produtos' },
    { rota: '/clientes', icone: 'people', rotulo: 'Clientes' },
    { rota: '/notas', icone: 'receipt_long', rotulo: 'Notas' },
  ]

  readonly mobile = signal<boolean>(this.isMobile())

  @HostListener('window:resize')
  onResize() {
    this.mobile.set(this.isMobile())
  }

  isLoggedIn(): boolean {
    return this.auth.isLoggedIn()
  }

  getUserName(): string {
    const user = this.auth.getUser()
    return user?.nome || user?.email || ''
  }

  logout(): void {
    this.auth.logout()
    this.router.navigate(['/login'])
  }

  private isMobile(): boolean {
    return typeof window !== 'undefined' ? window.innerWidth < 820 : false
  }
}
