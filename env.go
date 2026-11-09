package main

import (
	"fmt"
	"strings"
)

type playConfig struct {
	Enabled            bool
	PackageName        string
	ServiceAccountJSON string
	PubSubTopic        string
	SkipRemote         bool
}

func configFromArgs(args map[string]interface{}) playConfig {
	env := envFromArgs(args)
	return playConfig{
		Enabled:            env["PLAY_ENABLED"] == "" || truthy(env["PLAY_ENABLED"]),
		PackageName:        strings.TrimSpace(env["PLAY_PACKAGE_NAME"]),
		ServiceAccountJSON: strings.TrimSpace(env["PLAY_SERVICE_ACCOUNT_JSON"]),
		PubSubTopic:        strings.TrimSpace(env["PLAY_PUBSUB_TOPIC"]),
		SkipRemote:         truthy(env["PLAY_SKIP_REMOTE_VERIFY"]),
	}
}

func envFromArgs(args map[string]interface{}) map[string]string {
	out := map[string]string{}
	raw, _ := args["context_env_vars"].(map[string]interface{})
	for k, v := range raw {
		out[k] = strings.TrimSpace(fmt.Sprint(v))
	}
	return out
}

func stringArgs(args map[string]interface{}) map[string]string {
	out := map[string]string{}
	for k, v := range args {
		switch n := v.(type) {
		case string:
			out[k] = n
		case fmt.Stringer:
			out[k] = n.String()
		}
	}
	return out
}

func boolArg(args map[string]interface{}, key string) bool {
	switch n := args[key].(type) {
	case bool:
		return n
	case string:
		return truthy(n)
	default:
		return false
	}
}

func truthy(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
