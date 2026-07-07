using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddKeywordsAndCronToBusca : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<string>(
                name: "CronExpression",
                table: "Buscas",
                type: "text",
                nullable: true);

            migrationBuilder.AddColumn<string[]>(
                name: "Keywords",
                table: "Buscas",
                type: "text[]",
                nullable: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "CronExpression",
                table: "Buscas");

            migrationBuilder.DropColumn(
                name: "Keywords",
                table: "Buscas");
        }
    }
}
