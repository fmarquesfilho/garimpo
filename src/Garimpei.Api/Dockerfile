# ── Build stage ──────────────────────────────────────────────────────────────
# Context: repo root (docker build -f src/Garimpei.Api/Dockerfile .)
FROM mcr.microsoft.com/dotnet/sdk:10.0-alpine AS build
WORKDIR /src

# Copy protos (needed by Garimpei.Protos)
COPY protos/ /protos/

# Restore (cacheable layer)
COPY src/Garimpei.sln src/Directory.Build.props src/global.json ./
COPY src/Garimpei.Api/Garimpei.Api.csproj Garimpei.Api/
COPY src/Garimpei.Application/Garimpei.Application.csproj Garimpei.Application/
COPY src/Garimpei.Domain/Garimpei.Domain.csproj Garimpei.Domain/
COPY src/Garimpei.Infrastructure/Garimpei.Infrastructure.csproj Garimpei.Infrastructure/
COPY src/Garimpei.Protos/Garimpei.Protos.csproj Garimpei.Protos/
RUN dotnet restore

# Build
COPY src/ .
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
