import { Routes } from '@angular/router'

export const routes: Routes = [
  {
    path: '',
    redirectTo: 'home',
    pathMatch: 'full',
  },
  {
    path: 'login',
    loadComponent: () =>
      import('./pages/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: 'home',
    loadComponent: () =>
      import('./pages/home/home.component').then((m) => m.HomeComponent),
  },
  {
    path: 'produtos',
    loadComponent: () =>
      import('./pages/produtos/produtos.component').then(
        (m) => m.ProdutosComponent,
      ),
  },
  {
    path: 'clientes',
    loadComponent: () =>
      import('./pages/clientes/clientes.component').then(
        (m) => m.ClientesComponent,
      ),
  },
  {
    path: 'notas',
    loadComponent: () =>
      import('./pages/notas/notas.component').then((m) => m.NotasComponent),
  },
  {
    path: 'notas/:id',
    loadComponent: () =>
      import('./pages/nota-detalhe/nota-detalhe.component').then(
        (m) => m.NotaDetalheComponent,
      ),
  },
  {
    path: '**',
    redirectTo: 'home',
  },
]
