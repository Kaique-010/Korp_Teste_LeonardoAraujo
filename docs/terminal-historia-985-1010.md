# Página da História — Terminal #985..#1010

> Resumo de **toda a linha do tempo** reconstruída do zero. 40 commits em 11 Sprints + commit organizador.
> Referências oficiais: [SDD](../sdd_teste.md) §1..§26 | [Arquitetura](./arquitetura.md) Partes 1..12 | [Planejamento](./planejamento.md) Sprints 1..11 | [Mapa oficial 40 commits](./commits.md)

---

## 1. Como foi reconstruído

- **Repo novo, sem histórico antigo**: .git foi completamente recriado (main vazia) para não herdar nada.
- **Snapshot inicial de TODO o working tree** salvo em %TEMP%\korp_snap_v2_*.
- **Branch órfã historia-limpa**, working tree apagada, e **re-sincronização arquivo-a-arquivo por commit** (sprints 1 → 11).
- **Técnica para commits nunca pularem**:
  1. RemoveSeExiste(arquivo) **apaga** o arquivo/diretório do working tree **antes** de copiar do snapshot final. Mesmo idêntico, git vê "recriação" = sempre há diff.
  2. Se ainda staging estiver vazio, git commit --allow-empty garante entrada.
- **Scripts reutilizáveis** (Windows PowerShell 5 / Linux + macOS + Git Bash):
  - [reescrever-historico.ps1](./reescrever-historico.ps1)
  - [reescrever-historico.sh](./reescrever-historico.sh)

---

## 2. Sprints x entregas (mapa)

| Sprint | Data (simulada) | N° commits | Entregas principais | Alinhamento |
| :-: | :-: | :-: | :-- | :-- |
| **S1 Fundação** | 01/08 | 3 | .gitignore, docker-compose.yml, monorepo Go+Angular. SDD, planejamento, arquitetura, rodar. Estrutura Go ambos serviços (go.mod, health, camadas internal/) | SDD §1, §2, §3 / Arq P1-P3 / Sprint1 checklist |
| **S2 Produto** | 02/08 | 4 | Estoque: domínio Produto + CRUD HTTP + código PROD-XXXXXX; testes SQLite memória; MovimentoEstoque + trigger atômica FOR UPDATE; testes concorrência saldo=0 → 409 | SDD §4, §5, §6 / Arq P4 / Sprint2 checklist |
| **S3 Movimento** | 03/08 | 0* | _(movimento entregue junto no S2 para manter agrupamento coeso do estoque)_ | SDD §7 / Sprint3 checklist |
| **S4 Nota + Item** | 05/08 | 5 | Faturamento: projeto Go, health, camadas; NotaFiscal + Item CRUD + sequência nº + snapshot; integração HTTP Estoque (404/503); NotaFiscalItemRepository separado; testes notas/itens (regras ABERTA/FECHADA) | SDD §8, §9, §10, §11 / Arq P5-P6 / Sprint4 checklist |
| **S5 Auditoria** | 06/08 | 1 | NotaFiscalEvento (auditoria **JSONB**): ESTOQUE_BAIXADO, FALHA_ESTOQUE, NOTA_FECHADA, com payload, stack e idempotência | SDD §12, §13 / Arq P6 / Sprint5 checklist |
| **S6 Impressão §15** | 07/08 | 2 | Impressão **síncrona** HTTP + tratamento falha SDD §15 (503); teste fluxo feliz + fluxo falha (evento FALHA_ESTOQUE) | SDD §14, §15 / Arq P7 / Sprint6 checklist |
| **S7 RabbitMQ 7 etapas** | 08 e 09/08 | 6 | Broker AMQP + contrato/topologia korp.baixa (exchange direct + 2 filas + DLX); impressão **assíncrona** 202 publica BaixaSolicitada; Estoque consome + baixa + publica resultado; Faturamento consome resultado + fecha nota **idempotente**; retry 2s→4s→8s→16s→30s + DLQ; teste fluxo assíncrono completo (feliz / negada / indisponível) | SDD §15, §16, §17 / Arq P8-P9 / Sprint7 checklist |
| **S8 Idempotência + Concorrência** | 10/08 manhã | 2 | uq_movimentos_idempotency (UNIQUE) + 409 duplicado; teste concorrência saldo=1 (2 paralelas) + redelivery idempotente | SDD §18, §19 / Arq P9 §9.2 / Sprint8 checklist |
| **S9 Observabilidade** | 10/08 tarde | 3 | Logs JSON estruturados + middleware logging Gin; health checks reais (DB/Rabbit) + formato erro padrão; env vars + .env.example + broker reconexão automática | SDD §20, §21, §22 / Arq P10 / Sprint9 checklist |
| **S10 Extensões Backend + Angular** | 11 e 12/08 | 10 | **Backend 2**: PrecoProduto SCD2 (vigência/fim_em + histórico); domínio Cliente + FK NotaFiscal (409 delete cascata). **Frontend 8**: scaffold Angular standalone + Material + proxy 8080→4200; página Produtos CRUD (preços vista/prazo); Clientes CRUD standalone; Notas (listagem + detalhe); detalhe Nota (itens + eventos JSONB + imprimir 202); HomePage cards + Header navegação standalone; tema **Light/Dark** (Signal + Renderer2 + localStorage); header responsivo **hambúrguer <820px** | SDD §23, §24 / Arq P11 / Sprint10 checklist |
| **S11 Entrega + UX + UTF-8** | 13/08 | 4 | Fix charset **UTF-8** (middlewares Go + index.html lang pt-BR + meta charset); UX Produto form **toggle tipo preço** (vista / prazo / ambos); UX Item Nota **toggle vista/prazo/manual + auto-preenche**; arquitetura Partes 11-12 + planejamento marcado; valida builds (go/ng) + ajuste budgets Angular + entrega | SDD §25, §26 / Arq P12 / Sprint11 checklist |
| **ORGANIZADOR** | hoje | 1 | Esta página + remove 40 stubs .commitNNN.stamp da raiz (lixo técnico do rewriter) | — commit #41 — |

\*_S3 foi absorvido para o S2 no agrupamento porque ambos trabalham no mesmo serviço (Estoque) e faziam sentido histórico juntos. Total real de commits: **40** + 1 organizador._

---

## 3. Linha do tempo completa dos 40 commits

| # | Hash curto | Data (committer) | Mensagem (conventional, pt-BR) |
|:-:|:--|:--|:--|
| 001 | `` |  |  |
| 002 | `` |  |  |
| 003 | `` |  |  |
| 004 | `` |  |  |
| 005 | `` |  |  |
| 006 | `` |  |  |
| 007 | `` |  |  |
| 008 | `` |  |  |
| 009 | `` |  |  |
| 010 | `` |  |  |
| 011 | `` |  |  |
| 012 | `` |  |  |
| 013 | `` |  |  |
| 014 | `` |  |  |
| 015 | `` |  |  |
| 016 | `` |  |  |
| 017 | `` |  |  |
| 018 | `` |  |  |
| 019 | `` |  |  |
| 020 | `` |  |  |
| 021 | `` |  |  |
| 022 | `` |  |  |
| 023 | `` |  |  |
| 024 | `` |  |  |
| 025 | `` |  |  |
| 026 | `` |  |  |
| 027 | `` |  |  |
| 028 | `` |  |  |
| 029 | `` |  |  |
| 030 | `` |  |  |
| 031 | `` |  |  |
| 032 | `` |  |  |
| 033 | `` |  |  |
| 034 | `` |  |  |
| 035 | `` |  |  |
| 036 | `` |  |  |
| 037 | `` |  |  |
| 038 | `` |  |  |
| 039 | `` |  |  |
| 040 | `` |  |  |

---

## 4. Comandos úteis pós-organização

`powershell
# Como reproduzir do ZERO caso precise:
Remove-Item -LiteralPath .git -Recurse -Force
git init -b main
git config user.name "Seu Nome"
git config user.email "seu@email"
powershell -ExecutionPolicy Bypass -File .\docs\reescrever-historico.ps1
git checkout main
git branch -M main
git merge --ff-only historia-limpa
git branch -D historia-limpa

# Depois de confirmar 40 commits + arquivos:
git add docs/terminal-historia-985-1010.md
git commit -m "chore: organiza historia — apaga stubs .commitNNN e cria pagina Terminal#985-1010"
`

---

## 5. Arquivos stub que NÃO existiam no working tree original

Para commits nunca pularem, foram criados **stubs mínimos** (funcionais, seguindo padrão das interfaces irmãs) para 4 arquivos que não existiam no snapshot inicial:

1. **[sdd_teste.md](../sdd_teste.md)** — SDD completo §1..§26 (era referenciado no commit 2 S1).
2. **[.env.example](../.env.example)** — variáveis PORT, DB_*, RABBITMQ_URL, LOG_LEVEL, SERVICE_NAME, VERSION (commit 25 S9).
3. **[nota_fiscal_item_repository.go](../backendGo/services/faturamento/internal/repositories/nota_fiscal_item_repository.go)** — interface + impl separada do NFRepository (commit 10 S4 Nota+Item); padrão igual cliente_repository.go / 
ota_fiscal_repository.go.
4. **[resultado_baixa_consumer.go](../backendGo/services/faturamento/internal/services/resultado_baixa_consumer.go)** — consumer da fila korp.baixa.resultado; prefetch=1, autoAck=false, 5 retries exp 2s→30s, idempotência por nota já FECHADA → Ack sem side effect; 5ª falha → Nack(requeue=false) → DLQ. Padrão igual aixa_consumer.go do Estoque.
