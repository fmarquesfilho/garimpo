# Requirements Document

## Introduction

Coupon Monitoring enables Garimpei to automatically discover newly published coupons across supported marketplaces (Shopee, Amazon, Mercado Livre) and generate real-time alerts to affiliate publishers. The system periodically collects coupon data from marketplace APIs, detects new or updated coupons by comparing against previously known coupons stored in BigQuery, and dispatches notifications through configured channels (Telegram, WhatsApp). This extends the existing product monitoring pipeline (scheduler → collector → BigQuery → alerter) with a coupon-specific collection and detection flow that respects each marketplace's API characteristics and rate limits.

## Glossary

- **Coupon**: A marketplace-issued discount artifact with a code or claimable link, a discount value (percentage or fixed amount), validity period, applicable categories or products, and a minimum spend threshold
- **Coupon_Collector**: A gRPC microservice (one per marketplace) responsible for fetching available coupons from a marketplace's affiliate API
- **Coupon_Snapshot**: A time-stamped record of all coupons fetched in a single collection cycle, stored in BigQuery for historical comparison
- **Coupon_Detector**: The subsystem that compares the current Coupon_Snapshot against the previous one to identify newly appeared, expired, or modified coupons
- **Coupon_Alert**: A notification dispatched when the Coupon_Detector identifies a coupon that meets the user's configured alert criteria
- **Alert_Rule**: A user-defined configuration specifying which coupon characteristics (minimum discount, category, marketplace) should trigger a Coupon_Alert
- **Coupon_Store**: The BigQuery table that persists coupon snapshots and their metadata for historical analysis and deduplication
- **Scheduler**: The existing Go gRPC service that dispatches periodic collection jobs on cron schedules
- **Alerter**: The existing Go gRPC service that sends notifications to Telegram and WhatsApp channels

## Requirements

### Requirement 1: Coupon Collection from Shopee

**User Story:** As an affiliate publisher, I want the system to periodically fetch available coupons from the Shopee Affiliate API, so that I have an up-to-date inventory of active Shopee coupons.

#### Acceptance Criteria

1. WHEN the Scheduler triggers a Shopee coupon collection job, THE Coupon_Collector SHALL fetch coupon data from the Shopee Affiliate API `productOfferV2` endpoint, paginating through all available results in pages of up to 500 items, and extracting coupon-related fields (voucher code, discount value, minimum spend, validity period, applicable categories)
2. WHEN the Scheduler triggers a Shopee coupon collection job, THE Coupon_Collector SHALL authenticate using the tenant's Shopee credentials (AppID and Secret via HMAC-SHA256 signature as defined in the Shopee Affiliate API authentication scheme)
3. THE Coupon_Collector SHALL store each fetched coupon as a record in the Coupon_Store with fields: coupon_id, marketplace, code, discount_type (percentage or fixed), discount_value, min_spend, start_time, end_time, applicable_categories, collected_at
4. WHILE fetching coupons from Shopee, THE Coupon_Collector SHALL respect a minimum delay of 200 milliseconds between consecutive API requests to remain within rate limits
5. IF the Shopee API returns an error or the request does not complete within 30 seconds, THEN THE Coupon_Collector SHALL log the failure with the tenant UID and error details, and retry up to 2 additional attempts with a 5-second backoff between retries
6. IF all retry attempts fail, THEN THE Coupon_Collector SHALL skip the current collection cycle for that tenant and record a failed_collection event in the logs including the tenant UID, timestamp, and last error received
7. WHEN the Coupon_Collector completes a collection cycle for a tenant, THE Coupon_Collector SHALL mark any coupon in the Coupon_Store whose end_time is in the past as expired, so that only active coupons are presented to the publisher

### Requirement 2: Coupon Collection from Amazon

**User Story:** As an affiliate publisher, I want the system to periodically fetch promotional offers and coupons from the Amazon Creators API, so that I can monitor Amazon discount opportunities.

#### Acceptance Criteria

1. WHEN the Scheduler triggers an Amazon coupon collection job, THE Coupon_Collector SHALL fetch promotional offer data from the Amazon Creators API, extracting fields: offer_id, discount_type, discount_value, eligible_categories, start_date, end_date, and claiming_url
2. WHEN the Scheduler triggers an Amazon coupon collection job, THE Coupon_Collector SHALL authenticate using the tenant's Amazon OAuth 2.0 credentials, refreshing the access token using the stored refresh token if the current token is expired
3. THE Coupon_Collector SHALL normalize Amazon promotional offers into the Coupon_Store schema by mapping: offer_id to coupon_id, discount_type to discount_type, discount_value to discount_value, eligible_categories to applicable_categories, start_date to start_time, end_date to end_time, claiming_url to code, and setting marketplace to "amazon"
4. WHILE fetching coupons from Amazon, THE Coupon_Collector SHALL limit requests to a maximum of 1 request per second per tenant to comply with Amazon API rate limits
5. IF the Amazon API returns an HTTP 429 response, THEN THE Coupon_Collector SHALL wait 60 seconds before retrying, up to a maximum of 2 retry attempts
6. IF the Amazon API returns a server error (HTTP 5xx) or the request times out after 30 seconds, THEN THE Coupon_Collector SHALL log the failure with the tenant UID and error details, and retry up to 2 additional attempts with a 5-second backoff between retries
7. IF all retry attempts fail for a given Amazon collection request, THEN THE Coupon_Collector SHALL skip the current collection cycle for that tenant and record a failed_collection event in the logs
8. IF the tenant does not have Amazon credentials configured, THEN THE Coupon_Collector SHALL skip Amazon coupon collection for that tenant without logging an error
9. IF the Amazon OAuth 2.0 refresh token exchange fails, THEN THE Coupon_Collector SHALL log the authentication failure for the tenant and skip Amazon collection until credentials are re-authorized

### Requirement 3: Coupon Collection from Mercado Livre

**User Story:** As an affiliate publisher, I want the system to fetch available coupons and promotional campaigns from Mercado Livre, so that I can monitor discount opportunities on that marketplace.

#### Acceptance Criteria

1. THE Coupon_Collector SHALL fetch promotional deal data from the Mercado Livre Promotions API, extracting fields: deal_id, discount_type, discount_value, applicable_categories, start_date, end_date, and deal_url
2. WHEN the Scheduler triggers a Mercado Livre coupon collection job, THE Coupon_Collector SHALL authenticate using the tenant's Mercado Livre OAuth 2.0 access token, refreshing the token if expired
3. THE Coupon_Collector SHALL normalize Mercado Livre promotional deals into the unified Coupon_Store schema
4. IF the Mercado Livre access token is expired and the refresh token exchange fails, THEN THE Coupon_Collector SHALL log the authentication failure for the tenant and skip collection until credentials are re-authorized
5. IF the tenant does not have Mercado Livre credentials configured, THEN THE Coupon_Collector SHALL skip Mercado Livre coupon collection for that tenant without logging an error

### Requirement 4: Coupon Scheduling

**User Story:** As an affiliate publisher, I want coupon collection to run automatically on a configurable schedule, so that I receive timely updates without manual intervention.

#### Acceptance Criteria

1. THE Scheduler SHALL support registering coupon collection jobs with a configurable cron expression per tenant, with a default schedule of every 2 hours ("0 */2 * * *")
2. WHEN a coupon collection job is triggered, THE Scheduler SHALL dispatch collection requests to each marketplace Coupon_Collector for which the tenant has valid credentials
3. THE Scheduler SHALL execute marketplace coupon collections sequentially (Shopee, then Amazon, then Mercado Livre) for each tenant to avoid concurrent load spikes
4. WHEN a tenant configures a new Alert_Rule for coupons, THE Scheduler SHALL register a coupon collection job for that tenant if one does not already exist
5. IF a tenant disables all coupon Alert_Rules, THEN THE Scheduler SHALL pause the coupon collection job for that tenant but preserve the job configuration for reactivation

### Requirement 5: New Coupon Detection

**User Story:** As an affiliate publisher, I want the system to detect when new coupons appear on any marketplace, so that I am alerted to fresh opportunities as soon as they are discovered.

#### Acceptance Criteria

1. WHEN a new Coupon_Snapshot is stored, THE Coupon_Detector SHALL compare coupon IDs in the new snapshot against coupon IDs in the most recent previous snapshot for the same marketplace and tenant
2. WHEN a coupon ID appears in the current snapshot but not in the previous snapshot, THE Coupon_Detector SHALL classify that coupon as "newly_discovered"
3. WHEN a coupon ID appears in the previous snapshot but not in the current snapshot, THE Coupon_Detector SHALL classify that coupon as "expired_or_removed", provided the current snapshot contains at least 1 coupon; IF the current snapshot contains zero coupons, THEN THE Coupon_Detector SHALL skip classification for that cycle and log a warning indicating a potentially empty collection
4. WHEN a coupon exists in both snapshots but its discount_value or end_time has changed, THE Coupon_Detector SHALL classify that coupon as "modified"
5. WHEN detection completes for a snapshot, THE Coupon_Detector SHALL record each detection result in the Coupon_Store within 60 seconds of snapshot storage, including: coupon_id, marketplace, tenant owner_uid, detection_status (newly_discovered, expired_or_removed, modified), and detected_at timestamp
6. IF no previous Coupon_Snapshot exists for a marketplace-tenant combination, THEN THE Coupon_Detector SHALL classify all coupons in the first snapshot as "newly_discovered"
7. IF the Coupon_Detector fails to complete comparison for a snapshot due to a processing error, THEN THE Coupon_Detector SHALL log the failure with the snapshot ID and marketplace, discard any partial results, and retry detection once on the next scheduled collection cycle

### Requirement 6: Coupon Alert Rules

**User Story:** As an affiliate publisher, I want to define rules for which coupons should trigger an alert, so that I only receive notifications for coupons that match my publishing strategy.

#### Acceptance Criteria

1. THE System SHALL allow the user to create Alert_Rules specifying: minimum discount threshold (either a percentage from 1% to 99%, or a fixed monetary amount from 0.01 to 99999.99), target marketplaces (one or more of Shopee, Amazon, Mercado Livre), target categories (optional list of up to 10 categories), and notification channel (Telegram or WhatsApp)
2. THE System SHALL store Alert_Rules in PostgreSQL associated with the tenant's owner_uid
3. WHEN a coupon is classified as "newly_discovered" or "modified", THE Coupon_Detector SHALL evaluate it against all active Alert_Rules for that tenant
4. WHEN a newly_discovered or modified coupon matches an Alert_Rule (discount_value meets or exceeds the rule's minimum threshold for the same discount_type, marketplace matches, and category matches if specified), THE System SHALL dispatch a Coupon_Alert to the configured notification channel
5. THE System SHALL support a maximum of 20 active Alert_Rules per tenant
6. IF the user attempts to create an Alert_Rule that would exceed the maximum of 20 active rules for the tenant, THEN THE System SHALL reject the request and return an error message indicating the maximum number of active rules has been reached
7. IF a coupon matches multiple Alert_Rules for the same tenant and channel, THEN THE System SHALL send a single consolidated notification listing all matched rules rather than duplicate messages
8. WHERE the user has specified target categories in an Alert_Rule, THE System SHALL match coupons whose applicable_categories intersect with the rule's target categories
9. THE System SHALL allow the user to update, deactivate, and delete existing Alert_Rules
10. IF a coupon's discount_type does not match the Alert_Rule's discount threshold type (percentage vs fixed), THEN THE System SHALL skip that rule for the coupon without triggering an alert

### Requirement 7: Coupon Alert Notifications

**User Story:** As an affiliate publisher, I want to receive coupon alerts on my Telegram or WhatsApp channel with relevant details, so that I can quickly evaluate and share the coupon with my audience.

#### Acceptance Criteria

1. WHEN a Coupon_Alert is triggered, THE Alerter SHALL send a notification containing: coupon marketplace, discount description (e.g., "20% OFF" or "R$15 de desconto"), minimum spend (if applicable), validity period (start and end dates formatted as DD/MM/YYYY), applicable categories, and a claimable link or code
2. THE Alerter SHALL format Telegram notifications using Markdown with bold discount value, inline category tags, and a clickable link
3. THE Alerter SHALL format WhatsApp notifications as plain text with emoji indicators (🎟️ for coupon, ⏰ for expiry, 🏷️ for category)
4. WHILE the coupon has an end_time within 24 hours of the alert dispatch time, THE Alerter SHALL prepend an urgency indicator "⚡ Expira em breve!" to the notification message
5. IF the notification delivery to Telegram or WhatsApp fails, THEN THE Alerter SHALL retry delivery once after 30 seconds, and if the retry fails, log the failure with the alert_rule_id and coupon_id
6. THE Alerter SHALL include the tenant's affiliate link or tag in the coupon URL when the marketplace supports affiliate attribution on coupon links

### Requirement 8: Coupon Data Persistence

**User Story:** As an affiliate publisher, I want coupon history stored for analytics, so that I can identify patterns in coupon availability and optimize my publishing schedule.

#### Acceptance Criteria

1. THE Coupon_Store SHALL persist coupon records in a BigQuery table partitioned by `collected_at` date with fields: coupon_id, marketplace, code, discount_type, discount_value, min_spend, start_time, end_time, applicable_categories (JSON array), status (active, expired, claimed), owner_uid, and collected_at
2. THE Coupon_Store SHALL retain coupon records for a minimum of 90 days before any automated cleanup
3. WHEN the same coupon_id is collected in multiple cycles, THE Coupon_Store SHALL append a new record for each collection (append-only) rather than updating existing records, preserving full history
4. THE System SHALL support querying the Coupon_Store by marketplace, category, discount range, and time window for analytics dashboards
5. THE Coupon_Store SHALL store a `detection_status` field (newly_discovered, modified, expired_or_removed, unchanged) for each coupon record to enable detection analytics

### Requirement 9: Coupon Deduplication

**User Story:** As an affiliate publisher, I want the system to avoid sending duplicate alerts for the same coupon across consecutive collection cycles, so that my notification channels are not spammed with repeated information.

#### Acceptance Criteria

1. THE Coupon_Detector SHALL maintain a deduplication window of 24 hours per coupon_id per Alert_Rule, within which a coupon that was already classified as "newly_discovered" and alerted SHALL NOT trigger another alert for the same Alert_Rule, even if the coupon disappears and reappears within that window
2. IF a coupon classified as "modified" has already been alerted within the deduplication window, THEN THE Coupon_Detector SHALL send a new alert only if the discount_value increased compared to the previously alerted discount_value for that coupon_id and Alert_Rule combination
3. THE System SHALL persist the last alert timestamp and the alerted discount_value per coupon_id per Alert_Rule in PostgreSQL, retaining these records for at least 48 hours beyond the deduplication window to survive system restarts
4. WHEN the deduplication window expires for a coupon that is still active and still matches an Alert_Rule, THE System SHALL NOT re-alert unless the coupon's discount_value or end_time has changed compared to the values recorded at the last alert
5. IF a user modifies an existing Alert_Rule (changes minimum_discount, target marketplaces, or target categories), THEN THE Coupon_Detector SHALL reset the deduplication state for that Alert_Rule, allowing coupons that now match the updated criteria to trigger alerts in the next detection cycle
