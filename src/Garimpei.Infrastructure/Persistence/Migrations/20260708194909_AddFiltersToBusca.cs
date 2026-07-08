using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddFiltersToBusca : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<string[]>(
                name: "Categorias",
                table: "Buscas",
                type: "text[]",
                nullable: true);

            migrationBuilder.AddColumn<decimal>(
                name: "ComissaoMin",
                table: "Buscas",
                type: "numeric",
                nullable: true);

            migrationBuilder.AddColumn<string[]>(
                name: "Fontes",
                table: "Buscas",
                type: "text[]",
                nullable: true);

            migrationBuilder.AddColumn<int>(
                name: "VendasMin",
                table: "Buscas",
                type: "integer",
                nullable: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "Categorias",
                table: "Buscas");

            migrationBuilder.DropColumn(
                name: "ComissaoMin",
                table: "Buscas");

            migrationBuilder.DropColumn(
                name: "Fontes",
                table: "Buscas");

            migrationBuilder.DropColumn(
                name: "VendasMin",
                table: "Buscas");
        }
    }
}
