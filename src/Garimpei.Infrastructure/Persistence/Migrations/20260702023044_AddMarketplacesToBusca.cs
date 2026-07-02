using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddMarketplacesToBusca : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<string>(
                name: "Marketplaces",
                table: "Buscas",
                type: "text",
                nullable: false,
                defaultValue: "shopee");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "Marketplaces",
                table: "Buscas");
        }
    }
}
