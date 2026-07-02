-- Coupon snapshots: append-only log of collected coupons.
-- Partitioned by collected_at for cost-efficient time-range queries.
-- Partition expiration: 90 days (automatic cleanup by BigQuery).
CREATE TABLE IF NOT EXISTS `garimpo.coupon_snapshots` (
  coupon_id             STRING    NOT NULL,
  marketplace           STRING    NOT NULL,
  code                  STRING,
  discount_type         STRING    NOT NULL,
  discount_value        FLOAT64   NOT NULL,
  min_spend             FLOAT64,
  start_time            TIMESTAMP,
  end_time              TIMESTAMP,
  applicable_categories STRING,
  status                STRING    NOT NULL,
  detection_status      STRING,
  owner_uid             STRING    NOT NULL,
  collected_at          TIMESTAMP NOT NULL
)
PARTITION BY DATE(collected_at)
OPTIONS (
  partition_expiration_days = 90,
  description = "Append-only coupon snapshots for detection and analytics"
);
