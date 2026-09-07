# Data Model, Lexicons and XRPC

## What is a Lexicon?

A Lexicon is AT Protocol's schema language. It describes record structures, API endpoints and event stream messages.

```text
Lexicon → defines a record's structure
Lexicon → defines an XRPC endpoint's contract
XRPC    → lets clients call that endpoint
Record  → stores application data in a user's repository
```

For example, `app.bsky.feed.post` identifies a post record type. A simplified record looks like:

```json
{
  "$type": "app.bsky.feed.post",
  "text": "Hello AT Protocol!",
  "createdAt": "2026-09-07T10:00:00Z"
}
```

`$type` identifies the schema; `createdAt` is a datetime string. The Lexicon is the definition, while this JSON object is an instance of it.

### Why does AT Protocol need schemas?

Independent services need a shared understanding of data: a PDS hosts records, a Relay distributes repository updates, and an AppView interprets records for applications. Clients and Feed Generators also use defined API contracts.

Schemas make these contracts explicit, though each service still chooses which application types it supports. See the [Lexicon guide](https://atproto.com/guides/lexicon).

## NSID

NSID means **Namespaced Identifier**. It names a schema or API method, not an individual user's record.

```text
app.bsky.feed.post
└──┬───┘  │    └── name: post
   │      └─────── additional namespace: feed
   └────────────── reversed domain: bsky.app
```

NSIDs have at least three segments. The domain-based authority helps avoid naming collisions. For example, `dev.dayanch.blog.post` uses the authority `dayanch.dev`; publishing under that authority requires control of it. See the [NSID specification](https://atproto.com/specs/nsid).

## Records, Collections and Addresses

A repository groups records into collections identified by NSIDs. Each record has a record key (`rkey`) within its collection.

```text
at://<account-DID>/<collection-NSID>/<record-key>
```

An AT URI identifies the record's location in a repository. A CID identifies its encoded content: updating a record can keep its AT URI while changing its CID. See the [repository specification](https://atproto.com/specs/repository).

## Lexicon Validation

Validation checks required fields, types, formats and declared limits. For example, `"text": 42` fails a schema requiring a string. Optional fields may be omitted; accepting `null` requires a separate nullable declaration.

In Go, decoding JSON into a struct is only part of this work: schema constraints such as string length and datetime format still need validation. Passing validation does not prove authorship or grant permission to write a record.

## Custom Lexicons

Applications can define their own record types and APIs. This small learning example defines `dev.dayanch.blog.post`:

```json
{
  "lexicon": 1,
  "id": "dev.dayanch.blog.post",
  "defs": {
    "main": {
      "type": "record",
      "key": "tid",
      "record": {
        "type": "object",
        "required": ["text", "createdAt"],
        "properties": {
          "text": { "type": "string", "maxLength": 3000 },
          "createdAt": { "type": "string", "format": "datetime" }
        }
      }
    }
  }
}
```

`lexicon: 1` selects the schema-language version. `defs.main` is the primary definition, `key: "tid"` requires timestamp-based record keys, and string `maxLength` counts UTF-8 bytes. Defining a schema does not automatically implement an API or make Bluesky display its records. See the [Lexicon specification](https://atproto.com/specs/lexicon).

## XRPC

XRPC is AT Protocol's HTTP API convention. An endpoint uses `/xrpc/<NSID>` on the service implementing it.

For example, this query resolves a handle to a DID:

```http
GET /xrpc/com.atproto.identity.resolveHandle?handle=alice.example.com
```

### XRPC vs REST

| Style | Example | Organization |
| --- | --- | --- |
| REST | `GET /users/123` | Resources and HTTP methods |
| XRPC | `GET /xrpc/app.bsky.actor.getProfile?actor=alice.example.com` | Named methods and their schemas |

Both use HTTP; XRPC gives each operation an NSID.

### Query vs Procedure

| Lexicon type | Transport | Purpose | Example |
| --- | --- | --- | --- |
| `query` | HTTP GET | Read data | `app.bsky.actor.getProfile` |
| `procedure` | HTTP POST | Perform an action | `com.atproto.repo.createRecord` |
| `subscription` | WebSocket | Receive an event stream | `com.atproto.sync.subscribeRepos` |

Queries pass parameters in the URL. Procedures can accept a request body. Subscriptions use the separate [event stream protocol](https://atproto.com/specs/event-stream).

## Authentication

Some endpoints allow public reads; others require authentication and authorization.

- **OAuth:** the primary client authorization mechanism. An access token grants scoped access.
- **DPoP:** binds OAuth tokens to a client's key. Authenticated resource requests include a signed proof and use `Authorization: DPoP <access-token>`.
- **Legacy sessions:** `com.atproto.server.createSession` returns `accessJwt` and `refreshJwt`. Resource requests use `Authorization: Bearer <accessJwt>`; the refresh token renews the session through `com.atproto.server.refreshSession`.

The OAuth and legacy session flows have different token handling. See the [OAuth specification](https://atproto.com/specs/oauth) and [XRPC authentication documentation](https://atproto.com/specs/xrpc#authentication).

## Request / Response Encoding

JSON endpoints use `Content-Type: application/json` for their bodies. Lexicons declare supported encodings; some methods transfer binary data instead.

Errors use HTTP status codes and can include a JSON body such as:

```json
{
  "error": "InvalidRequest",
  "message": "Missing required parameter: actor"
}
```

Clients should check both the status and error body. See the [XRPC specification](https://atproto.com/specs/xrpc).

### API JSON vs Stored Data

The data model supports JSON and normalized CBOR representations. Repository records use DRISL-CBOR (the successor to DAG-CBOR) for reproducible encoding and content hashing.

- Supported values include strings, integers, booleans, arrays, objects, bytes and CID links; floating-point numbers are not allowed.
- JSON represents bytes as `{"$bytes": "aGVsbG8="}` and CID links as objects containing a `$link` string.
- Images and other files are blobs. Records contain blob references and metadata rather than embedding the whole file.

See the [data model specification](https://atproto.com/specs/data-model).
