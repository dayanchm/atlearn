# Poetry bot

The bot publishes a poem’s title, first four nonempty lines and author. SQL files are embedded in the Go binary, so no database or writable filesystem is needed.

## Selection and repeated requests

The dataset has 766 records. Each post contains the title, a blank line, the first four nonempty text lines, another blank line, and `— Author`. Blank lines and surrounding whitespace are removed; poems shorter than four lines use all available lines. SQL quoting is decoded by the loader. Excerpts exceeding the conservative 300-code-point or 3000-byte limit are skipped rather than cut mid-line. SQL records may contain extraction artifacts; editorial cleanup is a separate task.

Selection rotates through eligible authors every 20 minutes (every minute in local mode). Each author gets one turn per cycle, and their own poems advance on subsequent turns; authors with more records do not dominate consecutive posts. Each 20-minute interval uses a stable record key, so repeated or concurrent invocations within that interval try to create the same record instead of creating another post. If that record already exists, the PDS rejects the create and the endpoint returns an error. Check Function Logs and the account before retrying failed or timed-out requests. No local state is used, and missed intervals are not automatically backfilled. The endpoint remains `/api/daily` for compatibility with existing deployments.

Bluesky allows 300 graphemes and 3000 bytes. Counting code points is conservative: some otherwise valid emoji or combining-character text may be excluded. See the [post schema](https://github.com/bluesky-social/atproto/blob/main/lexicons/app/bsky/feed/post.json).

## Local development

Copy `.env.example` to `.env` and fill in the values. From `dailybot`, run:

```sh
go run .
```

The server listens on port 3000 unless `PORT` is set. The protected `/api/daily` endpoint accepts GET and POST for manual or scheduled invocation. An authenticated call publishes to the configured Bluesky account; visiting `/` only checks health.

Run automated tests without making real posts:

```sh
go test ./...
go vet ./...
```

### Publish every minute locally

Fill in `BSKY_HANDLE` and `BSKY_APP_PASSWORD` in `.env`, then run from `dailybot`:

```sh
go run . -local
```

The first real post is sent after one minute, followed by one post every minute while the process runs. Press Ctrl+C to stop. This mode does not start an HTTP server and does not require `CRON_SECRET` or an external scheduler. It uses minute-based poem selection and record keys; repeating a request in the same minute cannot create another record. Errors are logged and the next minute is attempted. The post still includes the title, first four nonempty lines and author.

Local posts end with `Çeşme: github.com/turkmenos/tm-data`. The source line is included in the length check; excerpts that no longer fit are skipped.
