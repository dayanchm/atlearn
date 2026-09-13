# 11. Firehose

## What is the ATProto firehose?

The firehose is a live stream of public ATProto activity. When someone creates a post, likes something, follows an account, or updates their repository, that change appears in the stream.

```text
user action -> PDS -> repository commit -> Relay -> firehose consumer
```

## Event streams and WebSockets

A normal HTTP request ends after one response. A firehose connection stays open and keeps delivering events. ATProto uses a binary WebSocket stream for this.

The main endpoint is:

```text
wss://<relay>/xrpc/com.atproto.sync.subscribeRepos
```

Messages use CBOR rather than JSON, so a client must decode them before reading their contents.

## Repository and commit events

Each account owns a repository on its PDS. A write creates a signed commit that points to the new repository state. A commit event includes the account DID, sequence number, revision, and operations such as `create`, `update`, or `delete`.

The operation path identifies the record type:

```text
app.bsky.feed.post/<rkey>    post
app.bsky.feed.like/<rkey>    like
app.bsky.graph.follow/<rkey> follow
```

## Reconnects, cursors, and backfill

WebSocket connections will eventually drop. Save the latest successfully processed sequence number and reconnect with it as the `cursor`. The Relay replays available missed events, then returns to the live stream.

```text
subscribe -> process event -> save cursor -> disconnect
                                      |
                                      +-> reconnect with cursor
```

This replay is called backfill. It has a limited retention window, so a very old cursor may require a full repository sync. Consumers should also tolerate duplicate events.

## Build

Start by connecting to `com.atproto.sync.subscribeRepos`, decoding each binary message, and printing its sequence number and type. Then inspect commit operations and keep only the collections you need:

```ts
const wanted = new Set([
  "app.bsky.feed.post",
  "app.bsky.feed.like",
  "app.bsky.graph.follow",
]);

for (const operation of commit.ops) {
  const collection = operation.path.split("/")[0];
  if (wanted.has(collection)) console.log(collection, operation.action);
}
```

Reconnect with a short increasing delay when the socket closes. Persist the cursor only after an event has been handled successfully.

# 12. Relay

## What is a Relay?

A Relay subscribes to many PDS instances, checks their repository updates, and combines them into one network-wide firehose. Applications can follow one Relay instead of connecting to every PDS.

## Discovery and synchronization

A Relay learns about PDS hosts from configured host lists, other Relays, and crawl requests. It then subscribes to each PDS event stream. For a new or outdated repository, it fetches repository data through the sync API and catches up before following live events.

```text
many PDSs -> validation and sync -> Relay -> one firehose
```

## Relay and PDS responsibilities

The PDS is the authority for its users' repositories. The Relay is a distributor: it mirrors public changes and makes them easier to consume, but it does not become the owner of that data.

## Validation, backfill, and recovery

Before forwarding data, a Relay checks content hashes, repository structure, commit history, and identity signatures. It is application-agnostic, so it does not decide whether a post is useful or how it should appear in a feed.

After downtime, the Relay resumes from saved sequence numbers. If the missed range is no longer available, it fetches the affected repository again and rebuilds its local state. Invalid or incomplete repositories are isolated and retried instead of silently trusted.

## Deep dive: Indigo Relay

The reference Relay lives in [`bluesky-social/indigo`](https://github.com/bluesky-social/indigo/tree/main/cmd/relay). A useful reading path is:

1. Start at `cmd/relay` to see service setup.
2. Follow the upstream subscription into the event handlers.
3. Find repository validation and persistence.
4. Follow the accepted event to `subscribeRepos`, where it is broadcast downstream.

Trace one post from its PDS commit, through validation and storage, to the event sent to a subscriber. That single path explains most of the Relay.

# 13. AppView

## What is an AppView?

An AppView reads repository events and builds views that an application can query quickly. Bluesky's AppView turns raw ATProto records into profiles, threads, feeds, search results, and notification data.

## Why is the PDS not enough?

A PDS stores canonical data for its own users. It is not designed to answer network-wide questions such as “Who liked this post?” or “What should appear in my timeline?” Those queries require data from many repositories.

## Indexing and hydration

Indexing extracts useful fields from firehose events and stores them in query-friendly tables or search indexes. Hydration takes a small result—often just post URIs—and adds the author, text, viewer state, media, and counts needed by the client.

```text
firehose -> index -> query -> hydrate -> API response
```

## Relationships, timelines, search, and counts

The AppView joins follow, like, reply, and repost records across users. From those relationships it can build timelines and conversation threads, power search, and calculate like, reply, and repost counts.

These values are derived views, not canonical records. They can briefly lag behind the PDS while new events are being indexed.

## Exercise

Explain the distinction in your own words:

> A PDS stores a user's canonical signed data. An AppView consumes that data from across the network and turns it into fast, application-oriented views.

## Sources

- [AT Protocol event stream specification](https://atproto.com/specs/event-stream)
- [Indigo Relay reference implementation](https://github.com/bluesky-social/indigo/tree/main/cmd/relay)
