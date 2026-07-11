using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class RemoveKeywordMigrateMarketplacesToArray : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            // 1. Convert existing comma-separated Marketplaces to JSON array before type change
            migrationBuilder.Sql("""
                UPDATE "Buscas"
                SET "Marketplaces" = (
                    SELECT jsonb_agg(trim(elem))
                    FROM unnest(string_to_array("Marketplaces", ',')) AS elem
                    WHERE trim(elem) != ''
                )::text
                WHERE "Marketplaces" IS NOT NULL AND "Marketplaces" != '';
                """);

            // 2. Default empty/null to ["shopee"]
            migrationBuilder.Sql("""
                UPDATE "Buscas"
                SET "Marketplaces" = '["shopee"]'
                WHERE "Marketplaces" IS NULL OR "Marketplaces" = '' OR "Marketplaces" = 'null';
                """);

            // 3. Drop legacy Keyword column (identity is now UUID + Keywords[])
            migrationBuilder.DropColumn(
                name: "Keyword",
                table: "Buscas");

            // 4. Change Marketplaces column type to jsonb
            migrationBuilder.AlterColumn<string>(
                name: "Marketplaces",
                table: "Buscas",
                type: "jsonb",
                nullable: false,
                oldClrType: typeof(string),
                oldType: "text");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AlterColumn<string>(
                name: "Marketplaces",
                table: "Buscas",
                type: "text",
                nullable: false,
                oldClrType: typeof(string),
                oldType: "jsonb");

            migrationBuilder.AddColumn<string>(
                name: "Keyword",
                table: "Buscas",
                type: "text",
                nullable: false,
                defaultValue: "");
        }
    }
}
