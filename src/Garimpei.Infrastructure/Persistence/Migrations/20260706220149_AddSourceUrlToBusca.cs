using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddSourceUrlToBusca : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<string>(
                name: "SourceUrl",
                table: "Buscas",
                type: "text",
                nullable: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "SourceUrl",
                table: "Buscas");
        }
    }
}
