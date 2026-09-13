package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type DescribeRepoResponse struct {
	Handle      string   `json:"handle"`
	DID         string   `json:"did"`
	Collections []string `json:"collections"`
}

type ListRecordsResponse struct {
	Records []Records `json:"records"`
	Cursor  string    `json:"cursor"`
}

type Records struct {
	URI   string          `json:"uri"`
	CID   string          `json:"cid"`
	Value json.RawMessage `json:"value"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: repository-inspector <handle-or-did>")
		os.Exit(1)
	}
	input := os.Args[1]

	baseURL := "https://phellinus.us-west.host.bsky.network"

	repo, err := describeRepo(baseURL, input)
	if err != nil {
		fmt.Println("describe repo:", err)
		os.Exit(1)
	}

	fmt.Println("Repository")
	fmt.Println("Handle:  %s\n", repo.Handle)
	fmt.Println("DID:  %s\n", repo.DID)

	fmt.Println()
	fmt.Println("Collections (%d)\n", len(repo.Collections))

	for _, collection := range repo.Collections {

		fmt.Printf("  %s\n", collection)
	}
	fmt.Println()

	for _, collection := range repo.Collections {
		fmt.Printf("Collection: %s\n", collection)
		records, err := listRecords(baseURL, repo.DID, collection)
		if err != nil {
			fmt.Printf("  error: %v\n\n", err)
			continue
		}
		fmt.Printf("  Records: %d\n", len(records.Records))
		for _, record := range records.Records {
			fmt.Printf("    URI: %s\n", record.URI)
			fmt.Printf("    CID: %s\n", record.CID)
			printRecordValue(record.Value)
			fmt.Println()
		}
		fmt.Println()
	}
}

func describeRepo(baseURL, repo string) (*DescribeRepoResponse, error) {

	endpoint := baseURL + "/xrpc/com.atproto.repo.describeRepo"
	params := url.Values{}
	params.Set("repo", repo)
	reqURL := endpoint + "?" + params.Encode()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}
	var result DescribeRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil

}

func listRecords(baseURL, repo, collection string) (*ListRecordsResponse, error) {

	endpoint := baseURL + "/xrpc/com.atproto.repo.listRecords"
	params := url.Values{}
	params.Set("repo", repo)
	params.Set("collection", collection)
	params.Set("limit", "10")
	reqURL := endpoint + "?" + params.Encode()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}
	var result ListRecordsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil

}

func printRecordValue(raw json.RawMessage) {

	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return
	}
	if recordType, ok := value["$type"].(string); ok {
		fmt.Printf("    Type: %s\n", recordType)
	}
	if text, ok := value["text"].(string); ok {
		text = strings.ReplaceAll(text, "\n", " ")
		if len(text) > 80 {
			text = text[:80] + "..."
		}
		fmt.Printf("    Text: %s\n", text)
	}
	if createdAt, ok := value["createdAt"].(string); ok {
		fmt.Printf("    Created: %s\n", createdAt)
	}

}
