"""Cupons: detection of new, modified, and expired coupons via BigQuery snapshot diff."""

import logging
from datetime import datetime
from typing import Optional

import httpx
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from config import settings
import bq_client

router = APIRouter(tags=["Cupons"])
logger = logging.getLogger(__name__)


class DetectCouponsRequest(BaseModel):
    owner_uid: str
    marketplace: str
    snapshot_timestamp: str  # RFC3339


class CouponDetectionItem(BaseModel):
    coupon_id: str
    detection_status: str  # newly_discovered, modified, expired_or_removed, unchanged
    discount_type: str
    discount_value: float
    end_time: Optional[str] = None
    applicable_categories: list[str] = []


class DetectCouponsResponse(BaseModel):
    owner_uid: str
    marketplace: str
    detections: list[CouponDetectionItem]
    total_new: int
    total_modified: int
    total_expired: int


@router.post("/detect-coupons", response_model=DetectCouponsResponse)
def detect_coupons(req: DetectCouponsRequest):
    """
    Compare coupon snapshots to detect newly_discovered, modified, and expired_or_removed coupons.
    After detection, POSTs results to C# API for alert evaluation.
    """
    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    # First check if current snapshot has any data (R5-AC3 safety check)
    count_sql = f"""
    SELECT COUNT(*) as cnt
    FROM {ds}.coupon_snapshots
    WHERE owner_uid = @owner_uid
      AND marketplace = @marketplace
      AND collected_at = TIMESTAMP(@current_ts)
    """

    from google.cloud.bigquery import ScalarQueryParameter

    count_rows = bq_client.query(count_sql, params=[
        ScalarQueryParameter("owner_uid", "STRING", req.owner_uid),
        ScalarQueryParameter("marketplace", "STRING", req.marketplace),
        ScalarQueryParameter("current_ts", "STRING", req.snapshot_timestamp),
    ])

    current_count = count_rows[0]["cnt"] if count_rows else 0
    if current_count == 0:
        logger.warning(
            "Empty coupon snapshot — skipping detection",
            extra={"owner_uid": req.owner_uid, "marketplace": req.marketplace},
        )
        return DetectCouponsResponse(
            owner_uid=req.owner_uid,
            marketplace=req.marketplace,
            detections=[],
            total_new=0,
            total_modified=0,
            total_expired=0,
        )

    # Main diff query
    detection_sql = f"""
    WITH current AS (
      SELECT coupon_id, discount_type, discount_value, end_time, applicable_categories
      FROM {ds}.coupon_snapshots
      WHERE owner_uid = @owner_uid
        AND marketplace = @marketplace
        AND collected_at = TIMESTAMP(@current_ts)
    ),
    previous AS (
      SELECT coupon_id, discount_type, discount_value, end_time, applicable_categories
      FROM {ds}.coupon_snapshots
      WHERE owner_uid = @owner_uid
        AND marketplace = @marketplace
        AND collected_at = (
          SELECT MAX(collected_at)
          FROM {ds}.coupon_snapshots
          WHERE owner_uid = @owner_uid
            AND marketplace = @marketplace
            AND collected_at < TIMESTAMP(@current_ts)
        )
    )
    SELECT
      c.coupon_id,
      c.discount_type,
      c.discount_value,
      CAST(c.end_time AS STRING) as end_time,
      c.applicable_categories,
      CASE
        WHEN p.coupon_id IS NULL THEN 'newly_discovered'
        WHEN c.discount_value != p.discount_value OR c.end_time != p.end_time THEN 'modified'
        ELSE 'unchanged'
      END AS detection_status
    FROM current c
    LEFT JOIN previous p ON c.coupon_id = p.coupon_id

    UNION ALL

    SELECT
      p.coupon_id,
      p.discount_type,
      p.discount_value,
      CAST(p.end_time AS STRING) as end_time,
      p.applicable_categories,
      'expired_or_removed' AS detection_status
    FROM previous p
    LEFT JOIN current c ON p.coupon_id = c.coupon_id
    WHERE c.coupon_id IS NULL
    """

    rows = bq_client.query(detection_sql, params=[
        ScalarQueryParameter("owner_uid", "STRING", req.owner_uid),
        ScalarQueryParameter("marketplace", "STRING", req.marketplace),
        ScalarQueryParameter("current_ts", "STRING", req.snapshot_timestamp),
    ])

    # If no previous snapshot exists, all are newly_discovered (R5-AC6)
    # The LEFT JOIN already handles this — previous will be empty

    detections: list[CouponDetectionItem] = []
    total_new = 0
    total_modified = 0
    total_expired = 0

    for row in rows:
        status = row["detection_status"]
        if status == "unchanged":
            continue  # Don't include unchanged in response

        categories = []
        if row.get("applicable_categories"):
            import json
            try:
                categories = json.loads(row["applicable_categories"])
            except (json.JSONDecodeError, TypeError):
                categories = []

        detections.append(CouponDetectionItem(
            coupon_id=row["coupon_id"],
            detection_status=status,
            discount_type=row.get("discount_type", "percentage"),
            discount_value=row.get("discount_value", 0),
            end_time=row.get("end_time"),
            applicable_categories=categories,
        ))

        if status == "newly_discovered":
            total_new += 1
        elif status == "modified":
            total_modified += 1
        elif status == "expired_or_removed":
            total_expired += 1

    # POST detection results to C# API for alert evaluation
    if detections:
        _notify_alert_matcher(req.owner_uid, req.marketplace, detections)

    return DetectCouponsResponse(
        owner_uid=req.owner_uid,
        marketplace=req.marketplace,
        detections=detections,
        total_new=total_new,
        total_modified=total_modified,
        total_expired=total_expired,
    )


def _notify_alert_matcher(owner_uid: str, marketplace: str, detections: list[CouponDetectionItem]):
    """POST detection events to C# API internal endpoint for alert evaluation."""
    api_url = settings.csharp_api_url if hasattr(settings, "csharp_api_url") else "http://garimpei-api-v2:8080"
    url = f"{api_url}/internal/coupon-alerts/evaluate"

    payload = {
        "ownerUid": owner_uid,
        "marketplace": marketplace,
        "detections": [d.model_dump() for d in detections],
    }

    try:
        with httpx.Client(timeout=10.0) as client:
            resp = client.post(url, json=payload)
            if resp.status_code == 200:
                result = resp.json()
                logger.info(
                    "Alert evaluation complete",
                    extra={"alerts_sent": result.get("alerts_sent", 0), "owner_uid": owner_uid},
                )
            else:
                logger.warning(
                    "Alert evaluation failed",
                    extra={"status": resp.status_code, "body": resp.text[:200]},
                )
    except Exception as e:
        logger.error(
            "Failed to notify alert matcher",
            extra={"error": str(e), "owner_uid": owner_uid},
        )
