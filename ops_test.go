package main

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestWebhookRenewAndRevoke(t *testing.T) {
	renew := rtdnBody(t, 2, "tok", "sku")
	got := webhookPlay(string(renew))
	if got["action"] != "grant" || got["external_id"] != "tok" || got["created_by"] != "rtdn" {
		t.Fatalf("%v", got)
	}
	revoke := rtdnBody(t, 13, "tok", "sku")
	got = webhookPlay(string(revoke))
	if got["action"] != "cancel" || got["end_access_now"] != true {
		t.Fatalf("%v", got)
	}
	if webhookPlay(`{"message":{"data":""}}`)["action"] != "noop" && webhookPlay(`{}`)["action"] != "noop" {
		t.Fatal("empty notification should be a no-op")
	}
}

func TestVerifySkipRemote(t *testing.T) {
	got := verifyPlay(t.Context(), playConfig{Enabled: true, SkipRemote: true, PackageName: "com.app"}, map[string]interface{}{
		"tenant_id":      "t1",
		"product_id":     "sku",
		"purchase_token": "tok",
	})
	if got["action"] != "grant" || got["status"] != "active" || got["external_id"] != "tok" {
		t.Fatalf("%v", got)
	}
	if got["manage_url"] == "" {
		t.Fatal("expected manage url")
	}
}

func rtdnBody(t *testing.T, kind int, token, sku string) []byte {
	t.Helper()
	inner, err := json.Marshal(map[string]any{
		"subscriptionNotification": map[string]any{
			"notificationType": kind,
			"purchaseToken":    token,
			"subscriptionId":   sku,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	outer, err := json.Marshal(map[string]any{
		"message": map[string]any{"data": base64.StdEncoding.EncodeToString(inner)},
	})
	if err != nil {
		t.Fatal(err)
	}
	return outer
}
