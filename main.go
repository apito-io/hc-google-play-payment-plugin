package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/apito-io/go-apito-plugin-sdk"
)

func main() {
	plugin := sdk.Init("hc-google-play-payment-plugin", "0.1.0", "")
	plugin.RegisterMutation("playInitialize",
		sdk.FieldWithArgs("String", "Google Play purchases start in the app", map[string]interface{}{
			"project_id": sdk.StringArg("Project id"),
		}),
		initializeResolver,
	)
	plugin.RegisterMutation("playVerify",
		sdk.FieldWithArgs("String", "Verify a Google Play purchase", map[string]interface{}{
			"project_id":     sdk.StringArg("Project id"),
			"tenant_id":      sdk.StringArg("Tenant id"),
			"product_id":     sdk.StringArg("Product id"),
			"purchase_token": sdk.StringArg("Purchase token"),
			"package_name":   sdk.StringArg("Package name"),
			"decorate":       sdk.BooleanArg("Return the manage link only"),
		}),
		verifyResolver,
	)
	plugin.RegisterMutation("playWebhook",
		sdk.FieldWithArgs("String", "Handle a Google Play developer notification", map[string]interface{}{
			"project_id": sdk.StringArg("Project id"),
			"raw_body":   sdk.StringArg("Raw webhook body"),
		}),
		webhookResolver,
	)
	plugin.RegisterMutation("playCancel",
		sdk.FieldWithArgs("String", "Cancel or revoke a Google Play subscription", map[string]interface{}{
			"project_id":        sdk.StringArg("Project id"),
			"product_id":        sdk.StringArg("Product id"),
			"purchase_token":    sdk.StringArg("Purchase token"),
			"mode":              sdk.StringArg("cancel or revoke"),
			"force_ledger_only": sdk.BooleanArg("Skip the Play API"),
		}),
		cancelResolver,
	)
	plugin.RegisterRESTAPI(sdk.GETEndpoint("/config", "Google Play configuration flags").Build(), configREST)
	plugin.Serve()
}

func initializeResolver(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("google play purchases start in the app")
}

func verifyResolver(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return reply(verifyPlay(ctx, configFromArgs(args), args))
}

func webhookResolver(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return reply(webhookPlay(stringArgs(args)["raw_body"]))
}

func cancelResolver(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return reply(cancelPlay(ctx, configFromArgs(args), stringArgs(args), boolArg(args, "force_ledger_only")))
}

func configREST(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	cfg := configFromArgs(args)
	return map[string]interface{}{
		"enabled":             cfg.Enabled,
		"package_name":        cfg.PackageName,
		"pubsub_topic":        cfg.PubSubTopic,
		"skip_remote_verify":  cfg.SkipRemote,
		"has_service_account": cfg.ServiceAccountJSON != "",
	}, nil
}

func reply(payload map[string]interface{}) (interface{}, error) {
	if payload == nil {
		payload = map[string]interface{}{"action": "noop"}
	}
	if msg, _ := payload["error"].(string); strings.TrimSpace(msg) != "" {
		return nil, fmt.Errorf("%s", msg)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}
