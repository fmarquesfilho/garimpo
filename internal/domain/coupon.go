package domain

// Coupon represents a marketplace coupon/voucher collected from affiliate APIs.
type Coupon struct {
	ID                   string   // unique coupon identifier from marketplace
	Marketplace          string   // "shopee", "amazon", "mercadolivre"
	Code                 string   // voucher code or claiming URL
	DiscountType         string   // "percentage" or "fixed"
	DiscountValue        float64  // e.g. 20.0 for 20% or 15.00 for R$15
	MinSpend             float64  // minimum purchase amount (0 = no minimum)
	StartTime            int64    // Unix timestamp
	EndTime              int64    // Unix timestamp
	ApplicableCategories []string // category IDs/names this coupon applies to
	Status               string   // "active", "expired", "claimed"
	OwnerUID             string   // tenant that collected this coupon
	CollectedAt          int64    // Unix timestamp of collection
}

// Discount type constants.
const (
	DiscountTypePercentage = "percentage"
	DiscountTypeFixed      = "fixed"
)

// Coupon status constants.
const (
	CouponStatusActive  = "active"
	CouponStatusExpired = "expired"
)

// DetectionStatus classifies a coupon after snapshot comparison.
type DetectionStatus string

const (
	DetectionNewlyDiscovered  DetectionStatus = "newly_discovered"
	DetectionModified         DetectionStatus = "modified"
	DetectionExpiredOrRemoved DetectionStatus = "expired_or_removed"
	DetectionUnchanged        DetectionStatus = "unchanged"
)

// CouponDetection is the result of comparing two snapshots.
type CouponDetection struct {
	CouponID      string
	Marketplace   string
	OwnerUID      string
	Status        DetectionStatus
	DiscountType  string
	DiscountValue float64
	EndTime       int64
	Categories    []string
	DetectedAt    int64 // Unix timestamp
}
