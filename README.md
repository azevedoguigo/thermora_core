#### Thermora Core (thermora-core)

Serviço responsável pelo **Control Plane** do ecossistema Thermora. Gerencia autenticação, usuários, equipes colaborativas, controle de acesso (RBAC) e o provisionamento de dispositivos IoT. 

#### 1. Stack Tecnológica

* **Linguagem:** Go (1.22+)
* **Framework HTTP:** Gin (github.com/gin-gonic/gin)
* **ORM / Banco de Dados:** GORM + PostgreSQL 16
* **Autenticação:** JWT (github.com/golang-jwt/jwt/v5) com assinatura assimétrica (RS256) ou chave secreta compartilhada (HS256)
* **Gerenciamento de Config:** Variáveis de ambiente (.env via godotenv ou viper)
* **Containerização:** Docker & Docker Compose

#### 2. Arquitetura e Padrões de Projeto

O serviço deve seguir uma **Arquitetura em Camadas Idiomática**, mantendo baixo acoplamento e facilitando a manutenção: 

```text
thermora-core/
├── cmd/
│   └── api/
│       └── main.go           # Entrypoint, injeção de dependências e setup de rotas
├── internal/
│   ├── config/               # Carregamento e validação de variáveis de ambiente
│   ├── database/             # Conexão com Postgres e configuração do GORM
│   ├── domain/               # Modelos/Structs de domínio e erros customizados
│   ├── handler/              # Controladores HTTP (bind de JSON, status codes, DTOs)
│   ├── service/              # Regras de negócio puras (sem dependência de HTTP ou GORM)
│   ├── repository/           # Implementação de persistência (única camada que importa GORM)
│   └── middleware/           # Auth JWT, CORS, Logger e recuperação de panics
├── migrations/               # Scripts SQL de controle de schema
├── .env.example              # Template de variáveis necessárias
├── Dockerfile
└── go.mod
```

#### Diretrizes Obrigatórias

1. **Isolamento do GORM:** Os pacotes handler e service **não** devem importar gorm.io/gorm. O repository deve receber e retornar structs puras ou erros padrão.
2. **Propagação de Contexto:** Toda chamada ao banco deve utilizar db.WithContext(ctx) repassando o context.Context da requisição HTTP.
3. **Sem Acesso Cruzado ao Banco:** Este serviço não acessa o banco de séries temporais (TimescaleDB) do thermora-telemetry.

#### 3. Modelagem de Dados (Entidades Principais)

#### User

* id (UUID, PK)
* name (String, obrigatório)
* email (String, único, obrigatório)
* password_hash (String, obrigatório, nunca retornar no JSON)
* created_at, updated_at

#### Team

* id (UUID, PK)
* name (String, obrigatório)
* owner_id (UUID, FK -> User.id)
* created_at, updated_at

#### TeamMember (Tabela Associativa)

* team_id (UUID, PK/FK -> Team.id)
* user_id (UUID, PK/FK -> User.id)
* role (Enum: ADMIN, VIEWER)
* joined_at (Timestamp)

#### Device (Sensor ESP32)

* id (UUID, PK)
* team_id (UUID, FK -> Team.id)
* name (String, ex: "Sensor Quarto Principal")
* api_key_hash (String — chave usada pelo ESP32 para autenticar no Broker MQTT)
* is_active (Boolean, default: true)
* created_at

#### 4. Requisitos Funcionais (Escopo do MVP)

#### Módulo 1: Autenticação & Usuários

* [ ] **Cadastro (POST /v1/auth/register):** Criar usuário com hash de senha seguro (bcrypt).
* [ ] **Login (POST /v1/auth/login):** Validar credenciais e emitir um **Access Token (JWT)**. 

  * *Requisito de Integração:* O payload do JWT deve conter sub (user_id) e a lista de team_ids aos quais o usuário pertence, pois o serviço thermora-telemetry validará esse mesmo token para liberar o WebSocket.
* [ ] **Perfil (GET /v1/users/me):** Retornar dados do usuário autenticado.

#### Módulo 2: Equipes & Convites

* [ ] **Criar Equipe (POST /v1/teams):** Cria a equipe e adiciona automaticamente o criador na TeamMember com role ADMIN.
* [ ] **Listar Minhas Equipes (GET /v1/teams):** Lista todas as equipes em que o usuário é ADMIN ou VIEWER.
* [ ] **Convidar Membro (POST /v1/teams/:id/members):** Apenas usuários com role ADMIN na equipe podem adicionar outros usuários (via e-mail) definindo a role (ADMIN ou VIEWER).
* [ ] **Remover Membro (DELETE /v1/teams/:id/members/:userId):** Apenas ADMIN pode remover membros (impedir que o owner_id seja removido).

#### Módulo 3: Gestão de Dispositivos (Sensores)

* [ ] **Registrar Dispositivo (POST /v1/teams/:id/devices):** Apenas ADMIN da equipe pode cadastrar um novo sensor. 

  * *Regra de Segurança:* Gerar uma api_key aleatória segura (ex: thm_live_...), salvar apenas o **hash** no banco e exibir a chave em texto plano **uma única vez** na resposta da criação.
* [ ] **Listar Dispositivos (GET /v1/teams/:id/devices):** Qualquer membro (ADMIN ou VIEWER) pode listar os sensores da equipe.
* [ ] **Revogar/Desativar Dispositivo (PATCH /v1/teams/:id/devices/:deviceId/status):** Apenas ADMIN pode ativar ou desativar um sensor.

#### Módulo 4: Endpoint Interno de Validação (Service-to-Service)

* [ ] **Validar Credencial do Sensor (POST /internal/devices/auth):** Rota protegida por uma chave interna (X-Internal-Secret) que será consultada pelo Broker MQTT / thermora-telemetry para verificar se o device_id + api_key do ESP32 são válidos e ativos.

#### 5. Critérios de Aceite (Definition of Done)

Para considerar uma tarefa concluída antes de abrir Pull Request: 

1. O código compila sem warnings e segue a formatação padrão (gofmt).
2. Todas as rotas protegidas validam o token JWT e verificam a permissão do usuário na equipe solicitada (evitar falha de IDOR).
3. Respostas de erro seguem um padrão JSON consistente (ex: {"error": "mensagem descritiva"}) com os HTTP Status Codes adequados (400, 401, 403, 404, 409).
4. Testes unitários escritos para as regras de negócio críticas na camada service (ex: bloqueio de convite por usuário VIEWER).
5. O serviço sobe corretamente via docker-compose conectando-se ao PostgreSQL.