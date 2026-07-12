using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddLojaEntity : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "Lojas",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    OwnerUid = table.Column<string>(type: "text", nullable: false),
                    ShopId = table.Column<long>(type: "bigint", nullable: false),
                    Nome = table.Column<string>(type: "character varying(200)", maxLength: 200, nullable: false),
                    NomeNormalizado = table.Column<string>(type: "character varying(200)", maxLength: 200, nullable: false),
                    Marketplace = table.Column<string>(type: "character varying(50)", maxLength: 50, nullable: false),
                    CronExpression = table.Column<string>(type: "text", nullable: true),
                    SourceUrl = table.Column<string>(type: "text", nullable: true),
                    OrigemPadrao = table.Column<string>(type: "text", nullable: true),
                    CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Lojas", x => x.Id);
                });

            migrationBuilder.CreateIndex(
                name: "IX_Lojas_OwnerUid",
                table: "Lojas",
                column: "OwnerUid");

            migrationBuilder.CreateIndex(
                name: "IX_Lojas_ShopId_Marketplace_OwnerUid",
                table: "Lojas",
                columns: new[] { "ShopId", "Marketplace", "OwnerUid" },
                unique: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "Lojas");
        }
    }
}
