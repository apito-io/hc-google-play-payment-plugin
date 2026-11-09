package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

type playSnapshot struct {
	Expiry string
}

type playClient struct {
	svc *androidpublisher.Service
}

func newPlayClient(ctx context.Context, serviceAccountJSON string) (*playClient, error) {
	raw := strings.TrimSpace(serviceAccountJSON)
	if raw == "" {
		return nil, fmt.Errorf("google play service account json required")
	}
	creds, err := google.CredentialsFromJSON(ctx, []byte(raw), androidpublisher.AndroidpublisherScope)
	if err != nil {
		return nil, fmt.Errorf("play credentials: %w", err)
	}
	svc, err := androidpublisher.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("androidpublisher service: %w", err)
	}
	return &playClient{svc: svc}, nil
}

func (c *playClient) get(ctx context.Context, packageName, subscriptionID, token string) (*playSnapshot, error) {
	if c == nil || c.svc == nil {
		return nil, fmt.Errorf("play client not configured")
	}
	resp, err := c.svc.Purchases.Subscriptions.Get(strings.TrimSpace(packageName), strings.TrimSpace(subscriptionID), strings.TrimSpace(token)).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	snap := &playSnapshot{}
	if resp.ExpiryTimeMillis > 0 {
		snap.Expiry = time.UnixMilli(resp.ExpiryTimeMillis).UTC().Format(time.RFC3339)
	}
	return snap, nil
}

func (c *playClient) cancel(ctx context.Context, packageName, subscriptionID, token string) error {
	return c.svc.Purchases.Subscriptions.Cancel(strings.TrimSpace(packageName), strings.TrimSpace(subscriptionID), strings.TrimSpace(token)).Context(ctx).Do()
}

func (c *playClient) revoke(ctx context.Context, packageName, subscriptionID, token string) error {
	return c.svc.Purchases.Subscriptions.Revoke(strings.TrimSpace(packageName), strings.TrimSpace(subscriptionID), strings.TrimSpace(token)).Context(ctx).Do()
}

func alreadyCanceled(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "subscription is canceled") ||
		strings.Contains(msg, "already canceled") ||
		strings.Contains(msg, "already cancelled") ||
		(strings.Contains(msg, "404") && strings.Contains(msg, "subscription"))
}
