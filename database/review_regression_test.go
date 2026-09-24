package database

import (
	"context"
	"encoding/json"
	"testing"
)

func TestReviewCapabilitySnapshotsPurged(t *testing.T) {
	for _, bulk := range []bool{false, true} {
		db := newPromptPolicySQLiteTestDB(t)
		ctx := context.Background()
		id, err := db.InsertAccount(ctx, "snapshot", "test-rt", "")
		if err != nil {
			t.Fatal(err)
		}
		row, err := db.GetAccountByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.SaveModelCapabilities(ctx, ModelCapabilitySnapshot{AccountID: id, CredentialGeneration: row.CredentialGeneration, ObservedAt: 1, Models: map[string]map[string]json.RawMessage{"model": {"context_window": json.RawMessage("1000")}}}); err != nil {
			t.Fatal(err)
		}
		if _, err := db.conn.ExecContext(ctx, `UPDATE accounts SET status='deleted', deleted_at=CURRENT_TIMESTAMP WHERE id=$1`, id); err != nil {
			t.Fatal(err)
		}
		if bulk {
			_, err = db.PurgeDeletedAccounts(ctx)
		} else {
			err = db.PurgeAccount(ctx, id)
		}
		if err != nil {
			t.Fatal(err)
		}
		var count int
		if err := db.conn.QueryRowContext(ctx, `SELECT count(*) FROM model_capability_snapshots WHERE account_id=$1`, id).Scan(&count); err != nil || count != 0 {
			t.Fatalf("orphan snapshot remains: count=%d err=%v", count, err)
		}
	}
}
