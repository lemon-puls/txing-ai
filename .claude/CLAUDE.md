# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**txing-ai** is an AI chat platform built with Go (backend) and Vue 3 (frontend). It provides a unified interface for multiple AI providers (OpenAI, Volcengine, Polo) with features like WebSocket streaming, JWT authentication, intelligent load balancing, and a complete admin dashboard.

## Build Commands

```bash
# Generate Swagger docs and build everything
make all

# Quick build (skip code generation - debug only)
make build

# Generate Swagger API docs only
make gen

# Build frontend
make frontend

# Build for Windows
make win

# Build for Docker
make docker
```

## Running the Application

```bash
# Run directly (development)
go run cmd/main.go

# Run with custom port
go run cmd/main.go -port 8081

# The built binary
./output/txing-ai_<hash>
```

## Configuration

1. Copy `config.yaml.sample` to `runtime/config.yaml`
2. Configure MySQL, Redis, and other services
3. Frontend API client is auto-generated via OpenAPI

## Architecture Overview

### Backend Structure (Go/Gin)

```
cmd/main.go           - Application entry point
internal/
├── adapter/          - AI provider adapters (OpenAI, Volcengine, Polo, Eino)
├── agent/            - AI agent implementations (General, ToolCall, Resume, Travel)
├── app/              - Application lifecycle management
├── controller/       - HTTP request handlers (organized by domain)
├── middleware/        - HTTP middleware (auth, logging, recovery)
├── domain/           - Database models (GORM entities)
├── dto/               - Data Transfer Objects (request payloads)
├── vo/                - View Objects (response payloads)
├── enum/              - Enumerations
├── global/            - Shared configuration and constants
├── iface/             - Interface definitions
├── service/           - Business logic layer
├── route/             - Route registration
├── tool/              - Agent tools (file ops, web scraping, MCP integration)
└── utils/             - Utility functions
```

### Frontend Structure (Vue 3)

```
static/frontend/src/
├── api/               - API client (generated/ directory is auto-generated)
├── components/        - Reusable components
├── directives/        - Vue directives (v-permission for role-based access)
├── layouts/           - Page layouts
├── router/            - Vue Router configuration
├── stores/            - Pinia state management
├── styles/            - Global styles and Element Plus theme
└── views/             - Page components
```

## Key Patterns

### Adding a New Backend API

1. Create domain model in `internal/domain/`
2. Create DTO in `internal/dto/` and VO in `internal/vo/`
3. Create controller in `internal/controller/<domain>/controller.go`
4. Create route file in `internal/controller/<domain>/route.go`
5. Register route in `internal/route/route.go`
6. Add Swagger annotations and run `make gen`

### API Response Format

Use `internal/utils/httpresp.go` for consistent responses:
- Success: `httpresp.Success(c, data)`
- Error: `httpresp.Fail(c, code, message)`

### API Endpoint Conventions

- Create: `POST /api/users`
- Update: `PUT /api/users/{id}`
- Delete: `DELETE /api/users/{id}`
- List (paginated): `GET /api/users/list`
- Detail: `GET /api/users/{id}`

### Frontend API Integration

1. After backend API changes, regenerate Swagger docs: `make gen`
2. Start backend server
3. Generate frontend API client:
   ```bash
   cd static/frontend
   pnpm generate-api
   ```
4. Import and use in components:
   ```js
   import { defaultApi } from '@/api'
   const response = await defaultApi.apiAdminChannelListGet(page, size)
   ```

### Frontend Permission Directive

```vue
<!-- Role-based permission -->
<el-button v-permission:role="['admin']">Admin Only</el-button>

<!-- Hide when no permission -->
<el-button v-permission:role.hide="['admin']">Hidden for non-admin</el-button>
```

## AI Provider Adapters

The adapter system provides unified interface for multiple AI providers:

- `internal/adapter/openai/` - OpenAI-compatible API
- `internal/adapter/volcengine/` - Volcengine (ByteDance)
- `internal/adapter/polo/` - Polo API
- `internal/adapter/eino_openai/` - CloudWeGo Eino framework

Channel configuration includes priority, weight, model mappings for load balancing.

## Agent System

Located in `internal/agent/`:
- `base_agent.go` - Base agent implementation
- `general_agent.go` - General-purpose agent
- `toolcall_agent.go` - Agent with tool calling
- `resume_agent.go` - Resume assistant
- `travel_agent.go` - Travel assistant

Tools are in `internal/tool/` including file operations, web scraping, MCP integration.

## Database

- GORM for ORM
- Auto-migration on startup
- Models in `internal/domain/`

## Testing

```bash
# Run Go tests
go test ./...

# Run specific test
go test -v ./internal/tool/... -run TestWebScraping
```

## Code Style

### Language
- Think in English, respond in Chinese (when working with Chinese users)
- Keep technical terms in English
- Use bilingual comments

### Frontend
- Use Composition API with `<script setup>`
- Element Plus for UI components
- Pinia for state management
- Follow Vue 3 best practices

### Backend
- Idiomatic Go
- Swagger annotations for all APIs
- Proper error handling
- Use context for cancellation

## Background Admin Style

Admin pages use rounded style (buttons, cards) for visual consistency.
