# ── Build stage ──────────────────────────────────────────────────────────────
# Context: src/ (docker build -f Garimpei.Api/Dockerfile .)
FROM mcr.microsoft.com/dotnet/sdk:10.0-alpine AS build
WORKDIR /src

# Restore (cacheable layer) — only production projects
COPY Directory.Build.props global.json ./
COPY Garimpei.Api/Garimpei.Api.csproj Garimpei.Api/
COPY Garimpei.Application/Garimpei.Application.csproj Garimpei.Application/
COPY Garimpei.Domain/Garimpei.Domain.csproj Garimpei.Domain/
COPY Garimpei.Infrastructure/Garimpei.Infrastructure.csproj Garimpei.Infrastructure/
COPY Garimpei.Protos/Garimpei.Protos.csproj Garimpei.Protos/
RUN dotnet restore Garimpei.Api/Garimpei.Api.csproj

# Build (includes pre-generated proto stubs in Garimpei.Protos/Generated/)
COPY . .
RUN dotnet publish Garimpei.Api/Garimpei.Api.csproj -c Release -o /app --no-restore

# ── Runtime stage ────────────────────────────────────────────────────────────
FROM mcr.microsoft.com/dotnet/aspnet:10.0-alpine AS runtime
WORKDIR /app

RUN adduser -D -u 1001 appuser
USER appuser

COPY --from=build /app .

ENV ASPNETCORE_URLS=http://+:8080
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["dotnet", "Garimpei.Api.dll"]
