import { Component } from '@angular/core'
import { CommonModule } from '@angular/common'
import { FormsModule } from '@angular/forms'
import { Router } from '@angular/router'
import { AuthService } from '../../services/auth.service'

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.component.html',
  styleUrl: './login.component.scss',
})
export class LoginComponent {
  email = ''
  senha = ''
  loading = false
  erroMsg = ''

  constructor(
    private authService: AuthService,
    private router: Router,
  ) {}

  login(): void {
    this.loading = true
    this.erroMsg = ''

    this.authService.login(this.email, this.senha).subscribe({
      next: () => {
        this.loading = false
        this.router.navigate(['/home'])
      },
      error: (err) => {
        this.loading = false
        console.error('Erro no login:', err)
        this.erroMsg =
          err?.error?.error || 'Email ou senha inválidos. Tente novamente.'
      },
    })
  }
}
