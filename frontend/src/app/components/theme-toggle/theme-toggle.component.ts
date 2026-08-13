import { Component, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { ThemeService } from '../../services/theme.service';

@Component({
  selector: 'app-theme-toggle',
  standalone: true,
  imports: [MatButtonModule, MatIconModule, MatTooltipModule],
  template: `
    <button
      mat-icon-button
      [matTooltip]="theme.theme() === 'dark' ? 'Mudar para tema claro' : 'Mudar para tema escuro'"
      (click)="theme.alternar()"
      aria-label="Alternar tema claro/escuro"
    >
      @if (theme.theme() === 'dark') {
        <mat-icon>light_mode</mat-icon>
      } @else {
        <mat-icon>dark_mode</mat-icon>
      }
    </button>
  `,
})
export class ThemeToggleComponent {
  readonly theme = inject(ThemeService);
}
