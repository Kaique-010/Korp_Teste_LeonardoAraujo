import { Injectable } from '@angular/core'
import { HttpClient } from '@angular/common/http'
import { Observable, tap } from 'rxjs'

interface LoginRequest {
  email: string
  senha: string
}

interface User {
  id: number
  nome: string
  email: string
}

interface LoginResponse {
  access_token: string
  token_type: string
  user: User
}

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private readonly apiUrl = 'http://localhost:8083/auth'
  private readonly TOKEN_KEY = 'korp_token'
  private readonly USER_KEY = 'korp_user'

  constructor(private http: HttpClient) {}

  login(email: string, senha: string): Observable<LoginResponse> {
    const body: LoginRequest = { email, senha }

    return this.http
      .post<LoginResponse>(`${this.apiUrl}/login`, body)
      .pipe(tap((res) => this.saveSession(res)))
  }

  private saveSession(res: LoginResponse): void {
    localStorage.setItem(this.TOKEN_KEY, res.access_token)
    localStorage.setItem(this.USER_KEY, JSON.stringify(res.user))
  }

  getToken(): string | null {
    return localStorage.getItem(this.TOKEN_KEY)
  }

  getUser(): User | null {
    const raw = localStorage.getItem(this.USER_KEY)
    return raw ? (JSON.parse(raw) as User) : null
  }

  isLoggedIn(): boolean {
    return !!this.getToken()
  }

  logout(): void {
    localStorage.removeItem(this.TOKEN_KEY)
    localStorage.removeItem(this.USER_KEY)
  }
}
