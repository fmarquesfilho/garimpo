using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddAmazonCredentials : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<string>(
                name: "AmazonAccessKeyEnc",
                table: "TenantConfigs",
                type: "text",
                nullable: true);

            migrationBuilder.AddColumn<string>(
                name: "AmazonPartnerTag",
                table: "TenantConfigs",
                type: "text",
                nullable: true);

            migrationBuilder.AddColumn<string>(
                name: "AmazonSecretKeyEnc",
                table: "TenantConfigs",
                type: "text",
                nullable: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "AmazonAccessKeyEnc",
                table: "TenantConfigs");

            migrationBuilder.DropColumn(
                name: "AmazonPartnerTag",
                table: "TenantConfigs");

            migrationBuilder.DropColumn(
                name: "AmazonSecretKeyEnc",
                table: "TenantConfigs");
        }
    }
}
