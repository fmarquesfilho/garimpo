using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddLojaEnrichedFields : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<string>(
                name: "CoverUrl",
                table: "Lojas",
                type: "character varying(500)",
                maxLength: 500,
                nullable: true);

            migrationBuilder.AddColumn<string>(
                name: "Description",
                table: "Lojas",
                type: "character varying(2000)",
                maxLength: 2000,
                nullable: true);

            migrationBuilder.AddColumn<int>(
                name: "FollowerCount",
                table: "Lojas",
                type: "integer",
                nullable: true);

            migrationBuilder.AddColumn<string>(
                name: "ImageUrl",
                table: "Lojas",
                type: "character varying(500)",
                maxLength: 500,
                nullable: true);

            migrationBuilder.AddColumn<int>(
                name: "ItemCount",
                table: "Lojas",
                type: "integer",
                nullable: true);

            migrationBuilder.AddColumn<double>(
                name: "RatingStar",
                table: "Lojas",
                type: "double precision",
                nullable: true);

            migrationBuilder.AddColumn<string>(
                name: "ShopLocation",
                table: "Lojas",
                type: "character varying(200)",
                maxLength: 200,
                nullable: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "CoverUrl",
                table: "Lojas");

            migrationBuilder.DropColumn(
                name: "Description",
                table: "Lojas");

            migrationBuilder.DropColumn(
                name: "FollowerCount",
                table: "Lojas");

            migrationBuilder.DropColumn(
                name: "ImageUrl",
                table: "Lojas");

            migrationBuilder.DropColumn(
                name: "ItemCount",
                table: "Lojas");

            migrationBuilder.DropColumn(
                name: "RatingStar",
                table: "Lojas");

            migrationBuilder.DropColumn(
                name: "ShopLocation",
                table: "Lojas");
        }
    }
}
