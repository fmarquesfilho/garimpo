# Garimpei — Web App C# (ASP.NET Core 10)

Web App principal do Garimpei. Minimal API com Clean Architecture.

## Estrutura

```
src/
├── Garimpei.Api/             # Minimal API (endpoints, middleware, Program.cs)
├── Garimpei.Application/     # Casos de uso (MediatR handlers, validators)
├── Garimpei.Domain/          # Entidades, interfaces, value objects
├── Garimpei.Infrastructure/  # EF Core, gRPC clients, serviços externos
└── Garimpei.Protos/          # Referência aos .proto (gera client stubs)
```

## Pré-requisitos

- .NET SDK 10.0+
- PostgreSQL 17+ (ou `make up-deps` para Docker)
- Protobuf / buf (para geração de proto)

## Setup local

```bash
# Subir dependências
make up-deps

# Restore + build
cd src && dotnet restore && dotnet build

# Rodar
cd src/Garimpei.Api && dotnet run

# Ou via Docker Compose (tudo junto)
make up
```

## Endpoints

- `GET /` — status
- `GET /health` — health check
- `GET /openapi/v1.json` — OpenAPI spec
- `GET /api/v2/buscas` — listar buscas (autenticado)
- `GET /api/v2/curadoria/ranking` — ranking de produtos (autenticado)
