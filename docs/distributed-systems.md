# Core Concepts

`How does a system operate reliably when it involves multiple services, servers, and databases, as well as millions of events?`

## Eventual consistency

**When a piece of data changes,the entire system does not need to see the new data simultaneously.**

**For example**
````
User change profile
 ↓
PDS
 ↓
Relay
 ↓
AppView
````
A change may have occurred in the PDS, but AppView might only see it 300 ms later.

## Strong consistency
It is the expectation that if a write operation succeeds, any subsequent read operation will definitely see the new value.

````
balance = $100

write:

balance = $50

write operation succeeds

read:

→ definitely $50
`````

## Replication

Storing multiple copies of the same data.


````
          Database A
         /
Data ─── Database B
         \
          Database C
````

Replicas can keep data available when a server fails and can distribute read traffic. With asynchronous replication, a replica may temporarily return older data. Replication is not a backup: accidental deletions can also propagate to replicas.

## Partitioning

Splitting a large dataset into smaller parts called partitions.

For example, orders can be partitioned by month:

```text
Orders
 ├── January partition
 ├── February partition
 └── March partition
```

Partitions can make data easier to manage and allow queries to scan only the relevant parts. They do not have to live on separate servers.

## Sharding

Distributing horizontal partitions of data across separate database servers. Each shard holds a subset of the records.

```text
Users 1–1000    → Shard A
Users 1001–2000 → Shard B
Users 2001–3000 → Shard C
```

Sharding spreads storage and traffic across servers. The shard key determines where a record goes; a poor choice can overload one shard. Cross-shard queries and moving data between shards add complexity.

**Partitioning divides data; sharding distributes those divisions across servers. Replication copies data, and each shard can have its own replicas.**

## Idempotency

Performing the same operation multiple times has the same intended effect as performing it once.

```text
Set order status to "paid" twice → status is still "paid"
Increment a counter twice        → counter increases by 2
```

For operations such as creating a payment, a client can send an idempotency key. The server records the key and result so a retry can return the previous result without creating another payment. Concurrent requests with the same key must also be handled safely.

## Backpressure

A way for a slower consumer to make upstream producers slow down or stop sending work temporarily.

```text
Producer: 1000 events/second
Consumer:  100 events/second
              ↓
Buffer fills → signal producer to slow down
```

Bounded buffers and limits on in-flight work prevent unlimited memory growth. If producers cannot slow down, the system needs an explicit policy to reject, drop, or persist excess work.

## Retry strategies

Trying an operation again after a failure that may be temporary, such as a connection interruption or an unavailable service.

```text
Request → temporary failure
Wait
Retry   → success
```

Retry only suitable failures, set a maximum number of attempts, and respect an overall deadline. A timeout does not prove the original operation failed: the server may have completed it before the response was lost. Operations with side effects need idempotency or deduplication to make retries safe.

## Exponential backoff

A retry strategy that increases the waiting time after each failed attempt, usually up to a maximum delay.

```text
Failure 1 → wait 1 second
Failure 2 → wait 2 seconds
Failure 3 → wait 4 seconds
Failure 4 → wait 8 seconds
```

One formula is `delay = min(max_delay, initial_delay × 2^retry_index)`, with the first retry at index 0. Add randomness, called **jitter**, so many clients do not retry at exactly the same time. For example, choose a random wait between zero and the calculated delay.

## Circuit breakers

Temporarily stopping calls to a dependency when repeated failures suggest it is unhealthy.

```text
Closed    → requests pass through
Open      → requests fail quickly without calling the dependency
Half-open → allow a limited number of recovery probes
```

When failures cross a configured threshold, the circuit opens. After a waiting period, probes check whether the dependency has recovered. Successful probes close the circuit; failed probes reopen it. This gives the dependency time to recover and avoids wasting caller resources.

## Rate limiting

Restricting how much work a client or service can submit within a period of time.

```text
Limit: 100 requests per minute per user
Excess requests → reject or delay
```

A token bucket replenishes tokens at a fixed rate, and each request consumes a token. Its capacity allows a controlled burst. HTTP APIs often return `429 Too Many Requests` and may include `Retry-After` to indicate when to try again.

Rate limiting enforces a traffic policy; backpressure responds to a consumer's ability to keep up.

## Caching

Keeping a reusable copy of data in a faster or closer location to reduce latency and load on the original source.

```text
Request → Cache
           ├── Hit  → return cached value
           └── Miss → read database → populate cache → return value
```

Cached data can become stale. A time-to-live (TTL) limits how long an entry stays cached, while invalidation removes or updates entries when the source changes. If many requests miss the same entry at once, coalescing them into one fetch can prevent a surge of database reads.

## Queues

Holding work until consumers are ready to process it, allowing producers and consumers to run independently.

```text
Producer → Queue → Worker
                 → Worker
```

A worker acknowledges a message after processing it. With at-least-once delivery, an unacknowledged message can be delivered again, so processing must tolerate duplicates. Messages that repeatedly fail can be moved to a dead-letter queue for investigation and later replay.

A queue can absorb temporary traffic spikes, but a growing backlog means consumers are not keeping up. Durability depends on the queue's configuration, and multiple workers may finish tasks out of order.

## Streaming

Continuously processing events as they arrive, instead of waiting for a complete batch.

```text
User events → Event stream → Analytics consumer
                          → Notification consumer
```

In a log-based event stream, events are retained for a configured period, and independent consumers track their own positions, often called offsets. Consumers can replay retained events to rebuild state or recover from failures.

Ordering is commonly guaranteed within a partition, rather than across the entire stream. Stream processing must also account for duplicates and events that arrive late or out of order.

## Failure recovery

Restoring service and correct state after a process, server, or dependency fails.

```text
Worker processes an event
          ↓
Worker crashes before acknowledging it
          ↓
Queue delivers the event again
          ↓
Another worker processes it safely and acknowledges it
```

Recovery can involve restarting processes, failing over to replicas, restoring backups, or replaying events from a checkpoint. A checkpoint records progress so processing can resume without starting over.

Coordinate progress tracking with durable side effects, or use idempotency, so recovery does not silently skip work or repeat its effects. Test recovery procedures and define acceptable downtime and data loss; having a replica or backup alone does not guarantee successful recovery.
