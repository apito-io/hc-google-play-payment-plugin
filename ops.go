package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func verifyPlay(ctx context.Context, cfg playConfig, args map[string]interface{}) map[string]interface{} {
	fields := stringArgs(args)
	productID := strings.TrimSpace(fields["product_id"])
	pkg := firstNonEmpty(fields["package_name"], cfg.PackageName)
	if boolArg(args, "decorate") {
		return map[string]interface{}{
			"action":     "noop",
			"manage_url": manageURL(pkg, productID),
			"product_id": productID,
		}
	}
	if !cfg.Enabled {
		return map[string]interface{}{"error": "google play is disabled for this project"}
	}
	token := strings.TrimSpace(fields["purchase_token"])
	tenantID := strings.TrimSpace(fields["tenant_id"])
	if tenantID == "" || token == "" || productID == "" {
		return map[string]interface{}{"error": "tenant_id, purchase_token, and product_id required"}
	}
	periodEnd := ""
	if !cfg.SkipRemote {
		if pkg == "" {
			return map[string]interface{}{"error": "package_name required for google play verify"}
		}
		client, err := newPlayClient(ctx, cfg.ServiceAccountJSON)
		if err != nil {
			return map[string]interface{}{"error": err.Error()}
		}
		snap, err := client.get(ctx, pkg, productID, token)
		if err != nil {
			return map[string]interface{}{"error": "play verify: " + err.Error()}
		}
		periodEnd = snap.Expiry
	}
	return map[string]interface{}{
		"action":                  "grant",
		"status":                  "active",
		"tenant_id":               tenantID,
		"product_id":              productID,
		"external_id":             token,
		"current_period_end":      periodEnd,
		"match_tenant":            true,
		"issue_invoice":           true,
		"invoice_status":          "paid",
		"created_by":              "verify",
		"billing_contact_user_id": fields["billing_contact_user_id"],
		"manage_url":              manageURL(pkg, productID),
	}
}

func cancelPlay(ctx context.Context, cfg playConfig, fields map[string]string, force bool) map[string]interface{} {
	if !cfg.Enabled {
		return map[string]interface{}{"error": "google play is disabled for this project"}
	}
	mode := strings.ToLower(strings.TrimSpace(fields["mode"]))
	if mode == "" {
		mode = "cancel"
	}
	if mode != "cancel" && mode != "revoke" {
		return map[string]interface{}{"error": "mode must be cancel or revoke"}
	}
	productID := strings.TrimSpace(fields["product_id"])
	token := strings.TrimSpace(fields["purchase_token"])
	pkg := cfg.PackageName
	periodEnd := ""
	if !force && !cfg.SkipRemote {
		if token == "" || productID == "" {
			return map[string]interface{}{"error": "purchase token and product_id required to cancel on google play"}
		}
		if pkg == "" {
			return map[string]interface{}{"error": "package_name required in google play provider config"}
		}
		client, err := newPlayClient(ctx, cfg.ServiceAccountJSON)
		if err != nil {
			return map[string]interface{}{"error": err.Error()}
		}
		var playErr error
		if mode == "revoke" {
			playErr = client.revoke(ctx, pkg, productID, token)
		} else {
			playErr = client.cancel(ctx, pkg, productID, token)
		}
		if playErr != nil && !alreadyCanceled(playErr) {
			return map[string]interface{}{"error": "play cancel: " + playErr.Error()}
		}
		if snap, gerr := client.get(ctx, pkg, productID, token); gerr == nil && snap != nil {
			periodEnd = snap.Expiry
		}
	}
	return map[string]interface{}{
		"action":             "cancel",
		"external_id":        token,
		"product_id":         productID,
		"subscription_id":    fields["subscription_id"],
		"end_access_now":     mode == "revoke",
		"current_period_end": periodEnd,
		"created_by":         "cancel",
		"manage_url":         manageURL(pkg, productID),
	}
}

func webhookPlay(rawBody string) map[string]interface{} {
	note, err := decodeRTDN(rawBody)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	if note == nil {
		return map[string]interface{}{"action": "noop"}
	}
	token := note.Token
	productID := note.ProductID
	switch note.Type {
	case 1, 2, 4, 7:
		return map[string]interface{}{
			"action":              "grant",
			"status":              "active",
			"external_id":         token,
			"product_id":          productID,
			"issue_invoice":       true,
			"invoice_status":      "paid",
			"invoice_external_id": token + ":" + time.Now().UTC().Format("2006-01"),
			"invoice_note":        "rtdn_renewal",
			"created_by":          "rtdn",
		}
	case 3:
		return map[string]interface{}{
			"action":         "cancel",
			"external_id":    token,
			"product_id":     productID,
			"end_access_now": false,
			"created_by":     "rtdn",
		}
	case 12, 13:
		return map[string]interface{}{
			"action":         "cancel",
			"external_id":    token,
			"product_id":     productID,
			"end_access_now": true,
			"created_by":     "rtdn",
		}
	default:
		return map[string]interface{}{"action": "noop"}
	}
}

type rtdnNote struct {
	Type      int
	Token     string
	ProductID string
}

func decodeRTDN(rawBody string) (*rtdnNote, error) {
	body := []byte(strings.TrimSpace(rawBody))
	if len(body) == 0 {
		return nil, nil
	}
	var env struct {
		Message struct {
			Data string `json:"data"`
		} `json:"message"`
	}
	raw := body
	if err := json.Unmarshal(body, &env); err == nil && env.Message.Data != "" {
		decoded, err := base64.StdEncoding.DecodeString(env.Message.Data)
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(env.Message.Data)
			if err != nil {
				return nil, fmt.Errorf("invalid RTDN message data: %w", err)
			}
		}
		raw = decoded
	}
	var note struct {
		SubscriptionNotification *struct {
			NotificationType int    `json:"notificationType"`
			PurchaseToken    string `json:"purchaseToken"`
			SubscriptionID   string `json:"subscriptionId"`
		} `json:"subscriptionNotification"`
	}
	if err := json.Unmarshal(raw, &note); err != nil {
		return nil, fmt.Errorf("invalid developer notification: %w", err)
	}
	if note.SubscriptionNotification == nil {
		return nil, nil
	}
	return &rtdnNote{
		Type:      note.SubscriptionNotification.NotificationType,
		Token:     note.SubscriptionNotification.PurchaseToken,
		ProductID: note.SubscriptionNotification.SubscriptionID,
	}, nil
}

func manageURL(packageName, productID string) string {
	pkg := strings.TrimSpace(packageName)
	if pkg == "" {
		return ""
	}
	q := url.Values{}
	q.Set("package", pkg)
	if sku := strings.TrimSpace(productID); sku != "" {
		q.Set("sku", sku)
	}
	return "https://play.google.com/store/account/subscriptions?" + q.Encode()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
