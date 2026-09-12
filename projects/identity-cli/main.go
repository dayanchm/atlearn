package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type DIDDocument struct {
	ID          string    `json:"id"`
	AlsoKnownAs []string  `json:"alsoKnownAs"`
	Service     []Service `json:"service"`
}

type Service struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: identity <handle>")
		os.Exit(1)
	}

	handle := os.Args[1]

	did, err := resolveHandle(handle)
	if err != nil {
		fmt.Println("resolve handle", err)
		os.Exit(1)
	}

	doc, err := resolveDID(did)
	if err != nil {
		fmt.Println("resolve DID", err)
		os.Exit(1)
	}

	pds, err := findPDS(doc)
	if err != nil {
		fmt.Println("find PDS", err)
		os.Exit(1)
	}
	fmt.Printf("Handle: %s\n", handle)
	fmt.Println("DID: %s\n", did)
	fmt.Println("PDS: %s\n", pds)
}

func resolveHandle(handle string) (string, error) {
	txts, err := net.LookupTXT("_atproto." + handle)
	if err == nil {
		for _, txt := range txts {
			if strings.HasPrefix(txt, "did=") {
				return strings.TrimPrefix(txt, "did="), nil
			}
		}
	}
	return resolveHandleHTTP(handle)
}

func resolveHandleHTTP(handle string) (string, error) {
	url := "https://" + handle + "/.well-known/atproto-did"
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil {
		return "", err
	}
	did := strings.TrimSpace(string(buf[:n]))
	if !strings.HasPrefix(did, "did:") {
		return "", errors.New("invalid DID")
	}
	return did, nil
}

func resolveDID(did string) (*DIDDocument, error) {

	var url string
	switch {
	case strings.HasPrefix(did, "did:plc:"):
		url = "https://plc.directory/" + did
	case strings.HasPrefix(did, "did:web:"):
		domain := strings.TrimPrefix(did, "did:web:")
		url = "https://" + domain + "/.well-known/did.json"
	default:
		return nil, fmt.Errorf("unsupported DID method: %s", did)
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}
	var doc DIDDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil

}

func findPDS(doc *DIDDocument) (string, error) {

	for _, service := range doc.Service {
		if service.Type == "AtprotoPersonalDataServer" &&
			strings.HasSuffix(service.ID, "#atproto_pds") {
			return service.ServiceEndpoint, nil
		}
	}
	return "", errors.New("PDS not found")

}
