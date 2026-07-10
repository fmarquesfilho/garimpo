using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddBuscaShopNames : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<string>(
                name: "ShopNames",
                table: "Buscas",
                type: "jsonb",
                nullable: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "ShopNames",
                table: "Buscas");
        }
    }
}
