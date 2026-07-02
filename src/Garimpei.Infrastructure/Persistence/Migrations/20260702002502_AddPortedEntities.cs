using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddPortedEntities : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "Destinos",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    Nome = table.Column<string>(type: "text", nullable: false),
                    Tipo = table.Column<string>(type: "text", nullable: false),
                    Config = table.Column<string>(type: "text", nullable: false),
                    Ativo = table.Column<bool>(type: "boolean", nullable: false),
                    OwnerUid = table.Column<string>(type: "text", nullable: false),
                    CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Destinos", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Favoritos",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    ProdutoId = table.Column<string>(type: "text", nullable: false),
                    Nome = table.Column<string>(type: "text", nullable: false),
                    Preco = table.Column<decimal>(type: "numeric", nullable: false),
                    Comissao = table.Column<double>(type: "double precision", nullable: false),
                    Link = table.Column<string>(type: "text", nullable: true),
                    Imagem = table.Column<string>(type: "text", nullable: true),
                    Loja = table.Column<string>(type: "text", nullable: true),
                    Categoria = table.Column<string>(type: "text", nullable: true),
                    Origem = table.Column<string>(type: "text", nullable: true),
                    Ativo = table.Column<bool>(type: "boolean", nullable: false),
                    OwnerUid = table.Column<string>(type: "text", nullable: false),
                    CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Favoritos", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Publicacoes",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    ProdutoId = table.Column<string>(type: "text", nullable: false),
                    Nome = table.Column<string>(type: "text", nullable: false),
                    Categoria = table.Column<string>(type: "text", nullable: true),
                    Preco = table.Column<decimal>(type: "numeric", nullable: false),
                    Comissao = table.Column<double>(type: "double precision", nullable: false),
                    Link = table.Column<string>(type: "text", nullable: true),
                    Imagem = table.Column<string>(type: "text", nullable: true),
                    Estrategia = table.Column<string>(type: "text", nullable: true),
                    DestinoId = table.Column<string>(type: "text", nullable: true),
                    TemplateId = table.Column<string>(type: "text", nullable: true),
                    AgendadaEm = table.Column<DateTime>(type: "timestamp with time zone", nullable: true),
                    Status = table.Column<string>(type: "text", nullable: false),
                    Detalhe = table.Column<string>(type: "text", nullable: true),
                    OwnerUid = table.Column<string>(type: "text", nullable: false),
                    CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    EnviadaEm = table.Column<DateTime>(type: "timestamp with time zone", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Publicacoes", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Templates",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    Nome = table.Column<string>(type: "text", nullable: false),
                    Corpo = table.Column<string>(type: "text", nullable: false),
                    ComFoto = table.Column<bool>(type: "boolean", nullable: false),
                    Ativo = table.Column<bool>(type: "boolean", nullable: false),
                    OwnerUid = table.Column<string>(type: "text", nullable: false),
                    CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Templates", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "TenantConfigs",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    OwnerUid = table.Column<string>(type: "text", nullable: false),
                    Email = table.Column<string>(type: "text", nullable: true),
                    ShopeeAppId = table.Column<string>(type: "text", nullable: true),
                    ShopeeSecretEnc = table.Column<string>(type: "text", nullable: true),
                    TelegramTokenEnc = table.Column<string>(type: "text", nullable: true),
                    TelegramChatId = table.Column<string>(type: "text", nullable: true),
                    OnboardingStep = table.Column<int>(type: "integer", nullable: false),
                    AceitouTermos = table.Column<bool>(type: "boolean", nullable: false),
                    AceitouTermosEm = table.Column<DateTime>(type: "timestamp with time zone", nullable: true),
                    AlertaThreshold = table.Column<double>(type: "double precision", nullable: false),
                    AlertaApenasQuedas = table.Column<bool>(type: "boolean", nullable: false),
                    CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_TenantConfigs", x => x.Id);
                });

            migrationBuilder.CreateIndex(
                name: "IX_Destinos_OwnerUid",
                table: "Destinos",
                column: "OwnerUid");

            migrationBuilder.CreateIndex(
                name: "IX_Favoritos_OwnerUid",
                table: "Favoritos",
                column: "OwnerUid");

            migrationBuilder.CreateIndex(
                name: "IX_Favoritos_OwnerUid_ProdutoId",
                table: "Favoritos",
                columns: new[] { "OwnerUid", "ProdutoId" },
                unique: true);

            migrationBuilder.CreateIndex(
                name: "IX_Publicacoes_OwnerUid",
                table: "Publicacoes",
                column: "OwnerUid");

            migrationBuilder.CreateIndex(
                name: "IX_Publicacoes_Status",
                table: "Publicacoes",
                column: "Status");

            migrationBuilder.CreateIndex(
                name: "IX_Templates_OwnerUid",
                table: "Templates",
                column: "OwnerUid");

            migrationBuilder.CreateIndex(
                name: "IX_TenantConfigs_OwnerUid",
                table: "TenantConfigs",
                column: "OwnerUid",
                unique: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "Destinos");

            migrationBuilder.DropTable(
                name: "Favoritos");

            migrationBuilder.DropTable(
                name: "Publicacoes");

            migrationBuilder.DropTable(
                name: "Templates");

            migrationBuilder.DropTable(
                name: "TenantConfigs");
        }
    }
}
