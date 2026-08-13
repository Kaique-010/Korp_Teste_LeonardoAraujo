import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./pages/home/home.component').then((m) => m.HomeComponent),
    pathMatch: 'full',
  },
  {
    path: 'produtos',
    loadComponent: () =>
      import('./pages/produtos/produtos.component').then((m) => m.ProdutosComponent),
  },
  {
    path: 'clientes',
    loadComponent: () =>
      import('./pages/clientes/clientes.component').then((m) => m.ClientesComponent),
  },
  {
    path: 'notas',
    loadComponent: () =>
      import('./pages/notas/notas.component').then((m) => m.NotasComponent),
  },
  {
    path: 'notas/:id',
    loadComponent: () =>
      import('./pages/nota-detalhe/nota-detalhe.component').then((m) => m.NotaDetalheComponent),
  },
  { path: '**', redirectTo: '' },
];
