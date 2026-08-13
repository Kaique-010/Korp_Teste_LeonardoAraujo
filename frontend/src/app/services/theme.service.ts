import { Injectable, Renderer2, RendererFactory2, inject, signal } from '@angular/core';

export type ThemeMode = 'light' | 'dark';

const LS_KEY = 'korp.theme';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly rendererFactory = inject(RendererFactory2);
  private readonly renderer: Renderer2 = this.rendererFactory.createRenderer(null, null);

  readonly theme = signal<ThemeMode>(this.lerPreferencia());

  constructor() {
    this.aplicar(this.theme());
  }

  alternar(): void {
    const proximo: ThemeMode = this.theme() === 'dark' ? 'light' : 'dark';
    this.definir(proximo);
  }

  definir(modo: ThemeMode): void {
    this.theme.set(modo);
    localStorage.setItem(LS_KEY, modo);
    this.aplicar(modo);
  }

  private aplicar(modo: ThemeMode): void {
    const body = document.body;
    this.renderer.removeClass(body, 'theme-light');
    this.renderer.removeClass(body, 'theme-dark');
    this.renderer.addClass(body, `theme-${modo}`);
  }

  private lerPreferencia(): ThemeMode {
    try {
      const salvo = localStorage.getItem(LS_KEY);
      if (salvo === 'light' || salvo === 'dark') return salvo;
    } catch {
      /* localStorage indisponível */
    }
    if (typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
      return 'dark';
    }
    return 'light';
  }
}
