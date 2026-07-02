using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Garimpei.Infrastructure.Persistence.Migrations
{
    /// <inheritdoc />
    public partial class AddCouponAlertRulesAndHistory : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "CouponAlertHistories",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    OwnerUid = table.Column<string>(type: "text", nullable: false),
                    CouponId = table.Column<string>(type: "text", nullable: false),
                    AlertRuleId = table.Column<Guid>(type: "uuid", nullable: false),
                    AlertedDiscountValue = table.Column<double>(type: "double precision", nullable: false),
                    AlertedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    ExpiresAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_CouponAlertHistories", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "CouponAlertRules",
                columns: table => new
                {
                    Id = table.Column<Guid>(type: "uuid", nullable: false),
                    OwnerUid = table.Column<string>(type: "text", nullable: false),
                    DiscountType = table.Column<string>(type: "text", nullable: false),
                    MinDiscount = table.Column<double>(type: "double precision", nullable: false),
                    Marketplaces = table.Column<string>(type: "text", nullable: false),
                    Categories = table.Column<string>(type: "text", nullable: false),
                    Channel = table.Column<string>(type: "text", nullable: false),
                    IsActive = table.Column<bool>(type: "boolean", nullable: false),
                    CreatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false),
                    UpdatedAt = table.Column<DateTime>(type: "timestamp with time zone", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_CouponAlertRules", x => x.Id);
                });

            migrationBuilder.CreateIndex(
                name: "IX_CouponAlertHistories_CouponId_AlertRuleId_AlertedAt",
                table: "CouponAlertHistories",
                columns: new[] { "CouponId", "AlertRuleId", "AlertedAt" });

            migrationBuilder.CreateIndex(
                name: "IX_CouponAlertHistories_OwnerUid",
                table: "CouponAlertHistories",
                column: "OwnerUid");

            migrationBuilder.CreateIndex(
                name: "IX_CouponAlertRules_OwnerUid",
                table: "CouponAlertRules",
                column: "OwnerUid");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "CouponAlertHistories");

            migrationBuilder.DropTable(
                name: "CouponAlertRules");
        }
    }
}
