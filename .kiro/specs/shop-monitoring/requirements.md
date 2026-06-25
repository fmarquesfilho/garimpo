# Requirements Document

## Introduction

Intelligent Shop Monitoring enables the user (Mileny) to configure Shopee stores for automatic product monitoring directly from the `/lojas` page. The system detects new products and price variations (opportunities) over time by collecting shop catalogs via rotational sampling and comparing consecutive snapshots stored in BigQuery. Alerts are surfaced in the UI with visual badges indicating price drops and rises. Throttling ensures Shopee API rate limits are respected when monitoring multiple shops concurrently.

## Glossary

- **Monitor**: The backend subsystem responsible for scheduling and executing periodic shop product collections
- **Shop_Form**: The frontend component on `/lojas` that accepts a Shopee shop URL or numeric shop ID for configuration
- **Shopee_Collector**: The Go service that fetches products from a shop via the Shopee Affiliate `shopOfferV2` GraphQL endpoint
- **Rotation_Cursor**: A persisted offset indicating which pages of a shop catalog to fetch in the current collection cycle
- **Snapshot**: A time-stamped record of products and prices from a shop, stored in the BigQuery `snapshots` table
- **Price_Alert**: A computed record indicating that a product price changed beyond the configured threshold between consecutive snapshots
- **Variation_Threshold**: The minimum percentage of price change (absolute value) required to generate a Price_Alert (default 15%)
- **Rate_Limiter**: The subsystem that spaces HTTP requests to the Shopee API to remain within the ~10 requests/second limit
- **Busca**: An existing data model representing a saved search profile; shop monitoring creates a Busca with `shop_ids` populated
- **Alerts_Tab**: The "📉 Preços" tab on the `/lojas` page that displays price variation alerts

## Requirements

### Requirement 1: Shop Configuration via Direct Form

**User Story:** As Mileny, I want to add a Shopee store to monitor directly from the `/lojas` page by entering its URL or ID, so that I do not need to navigate to the curadoria page to configure shop monitoring.

#### Acceptance Criteria

1. THE Shop_Form SHALL accept a single text input that recognizes three formats as valid: a shop URL matching `https://shopee.com.br/{shop_slug}`, a shop URL matching `https://shopee.com.br/shop/{shop_id}`, or a raw numeric shop ID (1 to 15 digits)
2. WHEN a valid Shopee shop URL of the format `https://shopee.com.br/shop/{shop_id}` is submitted, THE Shop_Form SHALL extract the numeric shop ID directly from the URL path segment
3. WHEN a valid Shopee shop URL of the format `https://shopee.com.br/{shop_slug}` is submitted, THE System SHALL resolve the slug to a numeric shop ID via the Shopee affiliate API before creating the Busca
4. WHEN a valid shop ID is obtained, THE Monitor SHALL call `POST /api/buscas` to create a Busca with `shop_ids` containing the extracted ID, `estrategia` set to "nicho", `ativo` set to true, and `cron` set to "0 */4 * * *"
5. WHEN the Busca is created successfully, THE Scheduler SHALL register a periodic collection job using the cron expression stored in the Busca
6. IF the submitted value does not match a recognized Shopee shop URL pattern and is not a string of 1 to 15 numeric digits, THEN THE Shop_Form SHALL display an inline validation error below the input field indicating the accepted formats, and the error SHALL remain visible until the user modifies the input
7. IF the `POST /api/buscas` request fails or returns a non-success status, THEN THE Shop_Form SHALL display an inline error message indicating the shop could not be saved, and SHALL preserve the user's input in the field
8. IF a Busca with the same shop ID already exists and is active, THEN THE System SHALL return an error indicating the shop is already being monitored, and the Shop_Form SHALL display this information to the user
9. WHEN a shop is added successfully, THE Shop_Form SHALL clear the input field and the new shop SHALL appear in the shops list within 1 second without requiring a page reload

### Requirement 2: Rotational Catalog Sampling

**User Story:** As Mileny, I want the system to progressively cover the entire catalog of a monitored shop over multiple collection cycles, so that large shops with hundreds of products are fully indexed without exceeding API rate limits in a single cycle.

#### Acceptance Criteria

1. THE Shopee_Collector SHALL fetch a configurable number of pages per collection cycle, bounded between 1 and 10 pages of 50 products each (default: 2 pages)
2. THE Monitor SHALL persist a Rotation_Cursor per shop in BigQuery indicating the next page offset (integer starting at 1) to fetch in the subsequent cycle
3. WHEN a collection cycle starts for a shop that has no persisted Rotation_Cursor, THE Shopee_Collector SHALL begin fetching from page 1
4. WHEN a collection cycle starts for a shop that has an existing Rotation_Cursor, THE Shopee_Collector SHALL begin fetching from the page indicated by that cursor
5. WHEN the Shopee_Collector reaches the last page of the catalog (shopOfferV2 returns hasNextPage=false before the configured page limit is reached), THE Monitor SHALL reset the Rotation_Cursor to page 1
6. WHEN the Rotation_Cursor is reset to page 1, THE Monitor SHALL record a full_scan_complete timestamp (UTC) for that shop to indicate the entire catalog has been covered
7. THE Shopee_Collector SHALL request pages sequentially within a single shop collection and process monitored shops one at a time to avoid concurrent requests against the Shopee API
8. IF the shopOfferV2 API returns an error during a collection cycle, THEN THE Shopee_Collector SHALL stop pagination for that shop, preserve the current Rotation_Cursor unchanged, and continue to the next monitored shop
9. WHEN a collection cycle completes the configured number of pages without reaching the last catalog page, THE Monitor SHALL update the Rotation_Cursor to the next unread page offset for that shop

### Requirement 3: Price Variation Detection

**User Story:** As Mileny, I want the system to detect products whose prices changed significantly between collection cycles, so that I can identify opportunities (price drops) and market movements (price rises).

#### Acceptance Criteria

1. WHEN a new Snapshot is stored, THE Monitor SHALL compare product prices in the new Snapshot against the most recent previous Snapshot for the same shop
2. WHEN a product price differs by more than the Variation_Threshold (default 15%), THE Monitor SHALL generate a Price_Alert record containing the product ID, product name, previous price, current price, percentage variation, and detection timestamp
3. THE Monitor SHALL compute variation as `(current_price - previous_price) / previous_price`
4. IF the previous price of a product is zero, THEN THE Monitor SHALL skip variation computation for that product and not generate a Price_Alert
5. THE Monitor SHALL store Price_Alert records in BigQuery so they can be queried by the UI, returning a maximum of 500 results per query
6. WHERE the user has configured a custom Variation_Threshold (valid range: 1% to 99%), THE Monitor SHALL use the user-specified threshold instead of the default 15%
7. THE Monitor SHALL generate Price_Alert records for both price drops (negative variation) and price rises (positive variation) whose absolute value exceeds the threshold
8. IF no previous Snapshot exists for a shop, THEN THE Monitor SHALL skip price variation detection and not generate any Price_Alert records

### Requirement 4: Price Alerts UI

**User Story:** As Mileny, I want to see products with significant price changes on the `/lojas` page with clear visual indicators, so that I can quickly identify opportunities and act on them.

#### Acceptance Criteria

1. THE Alerts_Tab SHALL display a list of products with price variations exceeding the Variation_Threshold for the selected shop, showing a maximum of 50 alert rows
2. WHEN a price drop is detected, THE Alerts_Tab SHALL display a green badge with format "↓ X.X%" (1 decimal place) next to the product name
3. WHEN a price rise is detected, THE Alerts_Tab SHALL display a red badge with format "↑ X.X%" (1 decimal place) next to the product name
4. THE Alerts_Tab SHALL sort products by absolute variation percentage in descending order (largest changes first)
5. THE Alerts_Tab SHALL display the product name, previous price, current price, variation badge, and detection date (formatted as YYYY-MM-DD) for each alert row
6. WHEN no price variations exist for the selected shop, THE Alerts_Tab SHALL display the message "Nenhuma variação de preço detectada nos últimos N dias." where N is the configured lookback window (default 7)
7. WHEN the user clicks the "📤" button on an alert row, THE Alerts_Tab SHALL navigate to the `/publicar` page with the product data pre-filled
8. WHILE the Alerts_Tab is loading price variation data, THE Alerts_Tab SHALL display a loading indicator with text "Analisando variações…"
9. IF the price variation data request fails, THEN THE Alerts_Tab SHALL display an error message indicating the failure reason

### Requirement 5: Multi-Shop Throttling

**User Story:** As Mileny, I want the system to space out requests when monitoring multiple shops, so that the Shopee API rate limit (~10 req/s) is not exceeded and collections remain reliable.

#### Acceptance Criteria

1. WHILE collecting products from multiple shops in a single scheduled run, THE Rate_Limiter SHALL insert a minimum delay of 60 seconds between starting collection for consecutive shops
2. THE Rate_Limiter SHALL limit HTTP requests to the Shopee API to a maximum of 5 requests per second across all concurrent operations within a single process, counting each page-fetch or API call as one request
3. WHILE collecting pages within a single shop, THE Rate_Limiter SHALL insert a minimum delay of 200 milliseconds between consecutive page requests
4. IF a Shopee API request returns an HTTP 429 (rate limited) response, THEN THE Rate_Limiter SHALL wait 30 seconds before retrying the request, up to a maximum of 3 retry attempts for the same request
5. IF a Shopee API request fails after 3 retry attempts (due to HTTP 429 or network error), THEN THE Monitor SHALL log the failure including the shop ID and page number that failed, skip all remaining pages for the current shop, and proceed to the next shop in the queue
6. WHEN a scheduled collection run completes (whether all shops succeeded or some were skipped), THE Monitor SHALL log a structured entry containing: total elapsed duration in seconds, total number of Shopee API requests made, number of shops collected successfully, and number of shops skipped due to errors

### Requirement 6: Shop URL Parsing

**User Story:** As Mileny, I want to paste any Shopee shop link format and have the system extract the shop ID automatically, so that I do not need to manually find the numeric ID.

#### Acceptance Criteria

1. WHEN a URL matching the pattern `https://shopee.com.br/{shop_slug}` is provided (where `{shop_slug}` is a string of 1 to 100 alphanumeric characters, hyphens, or underscores that does not match another known Shopee path segment such as "shop", "product", or "m"), THE Shop_Form SHALL extract the slug and resolve it to a numeric shop ID via a lookup request to the Shopee API within 10 seconds
2. WHEN a URL matching the pattern `https://shopee.com.br/shop/{shop_id}` is provided (where `{shop_id}` is a numeric identifier of 5 to 15 digits), THE Shop_Form SHALL extract and use the numeric shop ID directly without an API lookup
3. WHEN a raw numeric string of 5 to 15 digits is provided (no URL prefix), THE Shop_Form SHALL treat the value directly as a shop ID
4. THE Shop_Form SHALL strip trailing slashes, query parameters (text after `?`), and URL fragments (text after `#`) from the input before parsing
5. IF the slug resolution API request fails, times out after 10 seconds, or returns no matching shop ID, THEN THE Shop_Form SHALL display an error message "Não foi possível encontrar essa loja. Verifique o link e tente novamente." and SHALL preserve the original input value in the field
6. IF the provided input does not match any supported format (slug URL, numeric URL, or raw numeric string of 5 to 15 digits), THEN THE Shop_Form SHALL reject the input and not add it to the monitored shops list
