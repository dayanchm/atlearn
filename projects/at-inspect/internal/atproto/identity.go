package atproto

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const appViewBase = "https://public.api.bsky.app"

func ResolveHandle(ctx context.Context, handle string) (string, error) {
	u := fmt.Sprintf("%s/xrpc/com.atproto.identity.resolveHandle?handle=%s",
		appViewBase, url.QueryEscape(handle))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resolveHandle: unexpected status %d", resp.StatusCode)
	}

	var out struct {
		Did string `json:"did"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Did, nil
}
