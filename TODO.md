# TelemetryGo — TODO

Checklist de itens pendentes para o projeto ficar completo.

---

## Backend

### Segurança

- [X] Corrigir `bootstrap.go:70-72` — `log.Fatalf` no erro de JWT (bloco vazio no momento)
- [X] Adicionar validação de `JWT_SECRET` no `.env.example`
- [X] Adicionar `binding:"required"` em todos os request DTOs (`user_create_request_dto.go`, `login_request_dto.go`)
- [X] Adicionar JSON tags nos DTOs (`user_create_request_dto.go`, `login_request_dto.go`) — campos não são populados corretamente pelo `ShouldBindJSON`
- [X] Adicionar middleware de rate limiting
- [X] Adicionar security headers (X-Content-Type-Options, X-Frame-Options, etc.)

### API

- [X] Implementar `/me` corretamente — buscar usuário no banco por ID e retornar dados reais (name, email, api_key)
- [X] Criar `FindById` no repository interface e implementações
- [x] Criar endpoint `PUT /api/v1/me` — atualizar perfil (nome, email)
- [x] Criar endpoint `POST /api/v1/auth/change-password` — verificar senha atual e atualizar
- [x] Criar endpoint `POST /api/v1/auth/rotate-api-key` — gerar nova API key
- [x] Criar endpoint `DELETE /api/v1/users/me` — deletar conta
- [x] Criar endpoint `POST /api/v1/auth/forgot-password` — solicitar redefinição de senha (retorna reset_token no body para demo)
- [x] Criar endpoint `POST /api/v1/auth/reset-password` — redefinir senha usando token JWT
- [x] Criar `DELETE /api/v1/events/:id` e `DELETE /api/v1/metrics/:id`
- [x] Adicionar paginação nos endpoints de listagem (`limit`, `offset` query params)
- [x] Adicionar filtros por `severity`, `type`, `service`, time range nos events
- [x] Adicionar filtros por `status`, `name`, `service` nos metrics
- [x] Adicionar handler 404 customizado (JSON response em vez do HTML default do Gin)

### Validação e Error Handling

- [X] Cassandra `FindAll` deve retornar erro ao client em vez de slice vazio silencioso
- [X] Event/Metric list controllers precisam propagar erros do repositório
- [X] Publish errors não devem ser ignorados — retornar erro ou incluir no response
- [X] Timestamps inválidos devem retornar erro ao client em vez de fallback para `time.Now()`
- [X] Tratar erro de email duplicado na criação de usuário com mensagem amigável
- [X] Corrigir typo `"unknowm"` → `"unknown"` em `jwt_service.go:55`
- [X] In-memory user repository precisa de `sync.Mutex` para thread safety
- [X] Adicionar validação de `service`, `message`, `type`, `name`, `value` nos DTOs de ingest
- [X] Adicionar validação de que `Metric.Value` é numérico

### Code Quality

- [X] Consistência de tipos: `Event.Id` e `Metric.Id` devem ser `uuid.UUID` em vez de `string`
- [X] `Metric.Value` deve ser `float64` em vez de `string`
- [X] Consolidar `NewUser()` e `CreateUser()` — `CreateUser` deve chamar `NewUser(uuid.New(), ...)`
- [X] Remover import duplicado em `user_repository_i.go` (alias `vo` desnecessário)
- [X] Mensagens de validação de senha em inglês (consistente com o resto do código)
- [X] Parâmetros hardcoded (port, DB pool, token expiration, argon2 params) → mover para env/config
- [X] Adicionar graceful shutdown (signal handling + `http.Server.Shutdown`)
- [X] Usar `slog` ou `zerolog` em vez de `log.Println` puro
- [X] Health check deve verificar conectividade de Postgres, Cassandra e Redis

### Infraestrutura

- [x] Criar `Dockerfile` de produção (multi-stage com `CGO_ENABLED=0`)
- [x] Criar `docker-compose.prod.yml` para deploy
- [X] Criar `.env.example` documentando todas as variáveis
- [x] Criar `.golangci.yml` com configuração de linter
- [x] Criar `Makefile` ou task runner para comandos comuns (run, test, lint, build)

### Testes

- [x] Unit tests para `EventUsecase` (Ingest, List, Subscribe)
- [x] Unit tests para `MetricUsecase` (Ingest, List, Subscribe)
- [x] Unit tests para `UserController`, `EventController`, `MetricController`
- [x] Unit tests para domain entities e value objects (`Email`, `Severity`, `MetricStatus`)
- [x] Unit tests para `RedisPublisher`
- [x] Unit tests para `PostgresUserRepository`
- [x] Unit tests para `CassandraEventRepository` e `CassandraMetricRepository`
- [x] Unit tests para in-memory repositories
- [x] Teste de SSE streaming (`/events/stream`, `/metrics/stream`)
- [x] Teste de health check com dependências indisponíveis
- [X] Teste de criação de usuário com email duplicado
- [x] Compile-time interface assertions para in-memory repositories
- [x] Testes de fluxo forgot-password / reset-password (usecases, controller, integração)
- [x] Teste de forgot-password com email inexistente (não revela se existe)
- [x] Teste de reset com token inválido/expirado e senha fraca

---

## Frontend

### Funcional

- [x] Conectar `loadOverview()`, `loadMetrics()`, `loadEvents()` no dashboard — chamar nos `useEffect`
- [x] Criar `middleware.ts`/`proxy.ts` para proteger rotas `/dashboard` (redirecionar para `/login` se não autenticado)
- [x] Persistir auth state (accessToken, user, apiKey) em localStorage/cookie
- [x] Criar ação `fetchUser` no auth-store para rehidratar state no page load
- [x] Implementar login error feedback — conectar o Alert que está permanentemente `hidden`
- [x] Implementar página de events com dados reais da API
- [x] Implementar página de metrics com dados reais da API
- [x] Implementar forgot-password — conectar ao backend
- [x] Implementar reset-password — conectar ao backend
- [x] Implementar settings page — mostrar perfil e API key (fica pendente: change password e rotacionar API key precisam de endpoints no backend)

### UX / Design

- [x] Sidebar deve destacar a rota ativa — usar `usePathname()` em vez de `isActive` hardcoded
- [x] Adicionar loading skeletons nas páginas do dashboard
- [x] Adicionar scroll horizontal na tabela de métricas (`overflow-x-auto`)
- [x] Sidebar mobile deve fechar após navegação
- [x] Definir chart tokens (`--chart-1` a `--chart-5`) com cores reais em vez de monochrome
- [x] Usar design tokens para cores de tendência em vez de `text-emerald-500`/`text-amber-500` hardcoded
- [x] Adicionar espaço entre valor e unidade de métrica ("42 ms" em vez de "42ms")
- [x] Criar página `not-found.tsx` customizada
- [x] Criar `error.tsx` para error boundary no dashboard
- [x] Criar `loading.tsx` para loading states por rota

### Acessibilidade

- [x] Adicionar `aria-label` nos forms (login, register, forgot-password, reset-password)
- [x] Adicionar `aria-current="page"` nos links do sidebar para a rota ativa
- [x] Adicionar `aria-label` no trigger do dropdown do usuário
- [x] Adicionar `<caption>` ou `aria-label` na tabela de métricas

### Limpeza

- [x] Remover `useAxios()` hook se não será usado (ou substituir o singleton `api`)
- [x] Remover componentes não utilizados: `Tabs`, `Textarea`, `Select`
- [x] Remover `js-cookie` do `package.json` se não será usado
- [x] Remover `console.log("reset token")` da página de reset-password
- [x] Remover/reativar o Alert oculto no login
- [x] Definir o que fazer com `setApiKey` — usar ou remover do store
