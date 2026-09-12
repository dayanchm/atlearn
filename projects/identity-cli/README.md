# Identity CLI

A small Go command-line tool that resolves an AT Protocol handle to a DID and finds the user's PDS (Personal Data Server) endpoint in the DID document.

```text
Handle → DID → DID document → PDS endpoint
```

## Requirements

- Go version specified in `go.mod`: `1.27.0`.
- An internet connection for DNS queries and HTTPS requests.

The project uses only the Go standard library.

## Usage

From the repository root, navigate to the project directory and run the tool with a handle:

```sh
cd projects/identity-cli
go run . alice.bsky.social
```

Pass the handle without `@` or `https://`. The tool expects exactly one argument.

To build an executable:

```sh
go build -o identity .
./identity alice.bsky.social
```

The tool prints the handle, resolved DID, and PDS endpoint. If resolution fails, it prints an error message and exits with status code `1`.

## How it works

1. **Resolve the handle:** Search the DNS TXT records at `_atproto.<handle>` for a value starting with `did=`.
2. **Fall back to HTTPS:** If DNS resolution produces no result, request `https://<handle>/.well-known/atproto-did`.
3. **Fetch the DID document:** Use `https://plc.directory/<did>` for `did:plc`, or `https://<domain>/.well-known/did.json` for `did:web`.
4. **Find the PDS endpoint:** In the DID document's `service` list, find an entry with type `AtprotoPersonalDataServer` and an ID ending in `#atproto_pds`, then return its `serviceEndpoint` value.

Each HTTPS request has a timeout of 5 seconds.

## Files

| File | Description |
| --- | --- |
| `main.go` | Command-line entry point, handle and DID resolution, and PDS endpoint lookup. |
| `go.mod` | Module name and Go version. |

## Current limitations

- `did:web` resolution handles only the direct domain form; it does not handle paths or encoded ports.
- Identity matching is not verified against the DID document's `id` and `alsoKnownAs` values.
- The HTTPS handle resolution response is read once, up to 512 bytes; responses delivered in chunks may be read incompletely.
- Retries and caching are not implemented.
- The DID and PDS output lines use format specifiers with `fmt.Println`, causing incorrect output formatting; `go vet` may report these lines.

In its current form, this project is a basic implementation for learning the identity resolution flow.
