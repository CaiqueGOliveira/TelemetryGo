# TelemetryGo — TODO

Checklist de itens pendentes para o projeto ficar completo.

---

## Backend

### Segurança

- [ ] Corrigir `bootstrap.go:70-72` — `log.Fatalf` no erro de JWT (bloco vazio no momento)
- [ ] Adicionar validação de `JWT_SECRET` no `.env.example`
- [ ] Adicionar `binding:"required"` em todos os request DTOs (`user_create_request_dto.go`, `login_request_dto.go`)
- [ ] Adicionar JSON tags nos DTOs (`user_create_request_dto.go`, `login_request_dto.go`) — campos não são populados corretamente pelo `ShouldBindJSON`
- [ ] Adicionar limites de batch size nos endpoints de ingest (ex: max 500 events por request)
- [ ] Adicionar middleware de CORS (`gin-contrib/cors`)
- [ ] Adicionar middleware de rate limiting
- [ ] Adicionar security headers (X-Content-Type-Options, X-Frame-Options, etc.)

### API

- [ ] Implementar `/me` corretamente — buscar usuário no banco por ID e retornar dados reais (name, email, api_key)
- [ ] Criar `FindById` no repository interface e implementações
- [ ] Criar endpoint `PUT /api/v1/me` — atualizar perfil (nome, email)
- [ ] Criar endpoint `POST /api/v1/auth/change-password` — verificar senha atual e atualizar
- [ ] Criar endpoint `POST /api/v1/auth/rotate-api-key` — gerar nova API key
- [ ] Criar endpoint `DELETE /api/v1/users/me` — deletar conta
- [ ] Criar `DELETE /api/v1/events/:id` e `DELETE /api/v1/metrics/:id`
- [ ] Adicionar paginação nos endpoints de listagem (`limit`, `offset` query params)
- [ ] Adicionar filtros por `severity`, `type`, `service`, time range nos events
- [ ] Adicionar filtros por `status`, `name`, `service` nos metrics
- [ ] Adicionar handler 404 customizado (JSON response em vez do HTML default do Gin)

### Validação e Error Handling

- [ ] Cassandra `FindAll` deve retornar erro ao client em vez de slice vazio silencioso
- [ ] Event/Metric list controllers precisam propagar erros do repositório
- [ ] Publish errors não devem ser ignorados — retornar erro ou incluir no response
- [ ] Timestamps inválidos devem retornar erro ao client em vez de fallback para `time.Now()`
- [ ] Tratar erro de email duplicado na criação de usuário com mensagem amigável
- [ ] Corrigir typo `"unknowm"` → `"unknown"` em `jwt_service.go:55`
- [ ] In-memory user repository precisa de `sync.Mutex` para thread safety
- [ ] Adicionar validação de `service`, `message`, `type`, `name`, `value` nos DTOs de ingest
- [ ] Adicionar validação de que `Metric.Value` é numérico

### Code Quality

- [ ] Consistência de tipos: `Event.Id` e `Metric.Id` devem ser `uuid.UUID` em vez de `string`
- [ ] `Metric.Value` deve ser `float64` em vez de `string`
- [ ] Consolidar `NewUser()` e `CreateUser()` — `CreateUser` deve chamar `NewUser(uuid.New(), ...)`
- [ ] Remover import duplicado em `user_repository_i.go` (alias `vo` desnecessário)
- [ ] Mensagens de validação de senha em inglês (consistente com o resto do código)
- [ ] Parâmetros hardcoded (port, DB pool, token expiration, argon2 params) → mover para env/config
- [ ] Adicionar graceful shutdown (signal handling + `http.Server.Shutdown`)
- [ ] Usar `slog` ou `zerolog` em vez de `log.Println` puro
- [ ] Health check deve verificar conectividade de Postgres, Cassandra e Redis

### Infraestrutura

- [ ] Criar `Dockerfile` de produção (multi-stage com `CGO_ENABLED=0`)
- [ ] Criar `docker-compose.prod.yml` para deploy
- [ ] Criar `.env.example` documentando todas as variáveis
- [ ] Criar `.golangci.yml` com configuração de linter
- [ ] Criar `Makefile` ou task runner para comandos comuns (run, test, lint, build)

### Testes

- [ ] Unit tests para `EventUsecase` (Ingest, List, Subscribe)
- [ ] Unit tests para `MetricUsecase` (Ingest, List, Subscribe)
- [ ] Unit tests para `UserController`, `EventController`, `MetricController`
- [ ] Unit tests para domain entities e value objects (`Email`, `Severity`, `MetricStatus`)
- [ ] Unit tests para `RedisPublisher`
- [ ] Unit tests para `PostgresUserRepository`
- [ ] Unit tests para `CassandraEventRepository` e `CassandraMetricRepository`
- [ ] Unit tests para in-memory repositories
- [ ] Teste de SSE streaming (`/events/stream`, `/metrics/stream`)
- [ ] Teste de health check com dependências indisponíveis
- [ ] Teste de criação de usuário com email duplicado
- [ ] Compile-time interface assertions para in-memory repositories

---

## Frontend

### Funcional

- [ ] Conectar `loadOverview()`, `loadMetrics()`, `loadEvents()` no dashboard — chamar nos `useEffect`
- [ ] Criar `middleware.ts` para proteger rotas `/dashboard` (redirecionar para `/login` se não autenticado)
- [ ] Persistir auth state (accessToken, user, apiKey) em localStorage/cookie
- [ ] Criar ação `fetchUser` no auth-store para rehidratar state no page load
- [ ] Implementar login error feedback — conectar o Alert que está permanentemente `hidden`
- [ ] Implementar página de events com dados reais da API
- [ ] Implementar página de metrics com dados reais da API
- [ ] Implementar forgot-password — conectar ao backend (precisa do endpoint no backend primeiro)
- [ ] Implementar reset-password — conectar ao backend (precisa do endpoint no backend primeiro)
- [ ] Implementar settings page — form de perfil, change password, mostrar/rotacionar API key

### UX / Design

- [ ] Sidebar deve destacar a rota ativa — usar `usePathname()` em vez de `isActive` hardcoded
- [ ] Adicionar loading skeletons nas páginas do dashboard
- [ ] Adicionar scroll horizontal na tabela de métricas (`overflow-x-auto`)
- [ ] Sidebar mobile deve fechar após navegação
- [ ] Definir chart tokens (`--chart-1` a `--chart-5`) com cores reais em vez de monochrome
- [ ] Usar design tokens para cores de tendência em vez de `text-emerald-500`/`text-amber-500` hardcoded
- [ ] Adicionar espaço entre valor e unidade de métrica ("42 ms" em vez de "42ms")
- [ ] Criar página `not-found.tsx` customizada
- [ ] Criar `error.tsx` para error boundary no dashboard
- [ ] Criar `loading.tsx` para loading states por rota

### Acessibilidade

- [ ] Adicionar `aria-label` nos forms (login, register, forgot-password, reset-password)
- [ ] Adicionar `aria-current="page"` nos links do sidebar para a rota ativa
- [ ] Adicionar `aria-label` no trigger do dropdown do usuário
- [ ] Adicionar `<caption>` ou `aria-label` na tabela de métricas

### Limpeza

- [ ] Remover `useAxios()` hook se não será usado (ou substituir o singleton `api`)
- [ ] Remover componentes não utilizados: `Tabs`, `Textarea`, `Select`
- [ ] Remover `js-cookie` do `package.json` se não será usado
- [ ] Remover `console.log("reset token")` da página de reset-password
- [ ] Remover/reativar o Alert oculto no login
- [ ] Definir o que fazer com `setApiKey` — usar ou remover do store
