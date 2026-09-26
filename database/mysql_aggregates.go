package database

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (db *DB) getTrafficSnapshotMySQL(ctx context.Context) (*TrafficSnapshot, error) {
	snapshot := &TrafficSnapshot{}
	query := `
		WITH per_second AS (
			SELECT
				FLOOR(UNIX_TIMESTAMP(created_at)) AS sec,
				COUNT(*) AS req_count,
				COALESCE(SUM(total_tokens), 0) AS token_count
			FROM usage_logs
			WHERE created_at >= UTC_TIMESTAMP(3) - INTERVAL 5 MINUTE
			  AND TRIM(COALESCE(internal_reason, '')) = ''
			GROUP BY sec
		)
		SELECT
			COALESCE(SUM(CASE WHEN sec >= UNIX_TIMESTAMP(UTC_TIMESTAMP(3) - INTERVAL 10 SECOND) THEN req_count ELSE 0 END), 0) / 10.0 AS qps,
			COALESCE(MAX(req_count), 0) AS qps_peak,
			COALESCE(SUM(CASE WHEN sec >= UNIX_TIMESTAMP(UTC_TIMESTAMP(3) - INTERVAL 10 SECOND) THEN token_count ELSE 0 END), 0) / 10.0 AS tps,
			COALESCE(MAX(token_count), 0) AS tps_peak
		FROM per_second
	`
	if err := db.conn.QueryRowContext(ctx, query).Scan(&snapshot.QPS, &snapshot.QPSPeak, &snapshot.TPS, &snapshot.TPSPeak); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (db *DB) getChartAggregationMySQL(ctx context.Context, start, end time.Time, bucketMinutes int, channel string) (*ChartAggregation, error) {
	if bucketMinutes < 1 {
		bucketMinutes = 5
	}
	channel = strings.TrimSpace(channel)

	channelClause := ""
	timelineArgs := []interface{}{db.timeArg(start), db.timeArg(end), bucketMinutes}
	if channel != "" {
		channelClause = " AND channel = $4"
		timelineArgs = append(timelineArgs, channel)
	}

	result := &ChartAggregation{}
	timelineQuery := `
		SELECT
			CAST(FLOOR(UNIX_TIMESTAMP(created_at) / ($3 * 60)) * ($3 * 60) AS SIGNED) AS bucket_epoch,
			COUNT(*) AS requests,
			COALESCE(AVG(duration_ms), 0) AS avg_latency,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COALESCE(SUM(reasoning_tokens), 0) AS reasoning_tokens,
			COALESCE(SUM(cached_tokens), 0) AS cached_tokens,
			COALESCE(SUM(CASE WHEN status_code >= 400 AND status_code < 500 THEN 1 ELSE 0 END), 0) AS errors_4xx,
			COALESCE(SUM(CASE WHEN status_code >= 500 AND status_code < 600 THEN 1 ELSE 0 END), 0) AS errors_5xx
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2
		  AND status_code <> 499
		  AND TRIM(COALESCE(internal_reason, '')) = ''` + channelClause + `
		GROUP BY bucket_epoch
		ORDER BY bucket_epoch`

	rows, err := db.conn.QueryContext(ctx, timelineQuery, timelineArgs...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var bucketEpoch int64
		var point ChartTimelinePoint
		if err := rows.Scan(
			&bucketEpoch,
			&point.Requests,
			&point.AvgLatency,
			&point.InputTokens,
			&point.OutputTokens,
			&point.ReasoningTokens,
			&point.CachedTokens,
			&point.Errors4xx,
			&point.Errors5xx,
		); err != nil {
			_ = rows.Close()
			return nil, err
		}
		point.Bucket = time.Unix(bucketEpoch, 0).UTC().Format("2006-01-02T15:04:05Z")
		result.Timeline = append(result.Timeline, point)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	modelArgs := []interface{}{db.timeArg(start), db.timeArg(end)}
	modelChannelClause := ""
	if channel != "" {
		modelChannelClause = " AND channel = $3"
		modelArgs = append(modelArgs, channel)
	}
	modelQuery := `
		SELECT
			COALESCE(NULLIF(effective_model, ''), NULLIF(model, ''), 'unknown') AS model_name,
			COUNT(*) AS requests
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2
		  AND status_code <> 499
		  AND TRIM(COALESCE(internal_reason, '')) = ''` + modelChannelClause + `
		GROUP BY model_name
		ORDER BY requests DESC, model_name
		LIMIT 10`

	modelRows, err := db.conn.QueryContext(ctx, modelQuery, modelArgs...)
	if err != nil {
		return nil, err
	}
	for modelRows.Next() {
		var point ChartModelPoint
		if err := modelRows.Scan(&point.Model, &point.Requests); err != nil {
			_ = modelRows.Close()
			return nil, err
		}
		result.Models = append(result.Models, point)
	}
	if err := modelRows.Close(); err != nil {
		return nil, err
	}
	if err := modelRows.Err(); err != nil {
		return nil, err
	}

	if result.Timeline == nil {
		result.Timeline = []ChartTimelinePoint{}
	}
	if result.Models == nil {
		result.Models = []ChartModelPoint{}
	}
	return result, nil
}

func (db *DB) getMySQLAccountHealthBuckets(ctx context.Context, ids []int64, windowStart, now time.Time, blockCount int, bucketDuration time.Duration) (map[int64][]AccountHealthBucket, error) {
	args := []interface{}{db.timeArg(windowStart), db.timeArg(now), blockCount, bucketDuration.Seconds()}
	idFilter := ""
	if len(ids) > 0 {
		placeholders := make([]string, 0, len(ids))
		for _, id := range ids {
			args = append(args, id)
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		idFilter = " AND account_id IN (" + strings.Join(placeholders, ",") + ")"
	}

	query := `
		SELECT account_id, bucket_index,
			SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) AS success,
			SUM(CASE WHEN status_code < 200 OR status_code >= 300 THEN 1 ELSE 0 END) AS failed
		FROM (
			SELECT account_id, status_code,
				GREATEST(0, LEAST($3 - 1,
					CAST(FLOOR((UNIX_TIMESTAMP(created_at) - UNIX_TIMESTAMP($1)) / $4) AS SIGNED)
				)) AS bucket_index
			FROM usage_logs
			WHERE created_at >= $1 AND created_at <= $2
			  AND status_code <> 499 AND account_id > 0
			  AND ` + db.endUserUsageLogPredicate() + idFilter + `
		) AS scoped_logs
		GROUP BY account_id, bucket_index
		ORDER BY account_id, bucket_index`

	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]AccountHealthBucket)
	for rows.Next() {
		var accountID int64
		var bucketIndex, success, failed int
		if err := rows.Scan(&accountID, &bucketIndex, &success, &failed); err != nil {
			return nil, err
		}
		if bucketIndex < 0 || bucketIndex >= blockCount {
			continue
		}
		buckets := result[accountID]
		if buckets == nil {
			buckets = make([]AccountHealthBucket, blockCount)
			result[accountID] = buckets
		}
		buckets[bucketIndex] = AccountHealthBucket{Success: success, Failed: failed}
	}
	return result, rows.Err()
}
