# Daily poetry bot on Vercel

The bot serves `/api/daily` and publishes a poem’s title, first four nonempty lines and author every 20 minutes. SQL files are embedded in the Go binary, so no database or writable filesystem is needed.

## Deploy

1. Import this repository into Vercel and set **Root Directory** to `dailybot`.
2. Select the **Go** framework preset. Keep the default build settings; `vercel.json` configures the framework and 20-minute cron.
3. Add these environment variables to **Production** before deploying:

| Variable | Value |
| --- | --- |
| `BSKY_HANDLE` | Your Bluesky handle |
| `BSKY_APP_PASSWORD` | An app password for the bot account |
| `CRON_SECRET` | A long random secret; generate one with `openssl rand -hex 32` |
| `BSKY_PDS_URL` | Optional; defaults to `https://bsky.social`. Use your account's PDS for other hosts. |

4. Deploy to Production. In Settings → Cron Jobs, check `/api/daily`.
5. Visit `/` to check that the server runs. This health check does not publish anything.

Vercel runs the Go server using its `PORT` variable and the version in `go.mod`. See [Vercel's Go runtime documentation](https://vercel.com/docs/functions/runtimes/go).

Do not add `functions.main.go` to `vercel.json`: it triggers an API-directory function pattern check and prevents deployment. This project uses the Go server preset and the project's default function duration. If needed, set the duration to at least 60 seconds in Vercel's project settings to allow for the two outbound API requests.

The schedule is `*/20 * * * *`: at minutes 00, 20 and 40 of every hour (also in Ashgabat). **Vercel Pro or Enterprise is required for this built-in cron schedule. Hobby only allows daily jobs and will reject this configuration during deployment.** On Hobby, remove the `crons` entry and use an external scheduler to call the protected endpoint every 20 minutes with `Authorization: Bearer <CRON_SECRET>`. See [Vercel cron plan limits](https://vercel.com/docs/cron-jobs/usage-and-pricing).

Vercel sends that authorization header automatically for its own cron requests. Scheduled jobs run on production deployments. See [Vercel cron documentation](https://vercel.com/docs/cron-jobs/manage-cron-jobs).

## Selection and repeated requests

The dataset has 766 records. Each post contains the title, a blank line, the first four nonempty text lines, another blank line, and `— Author`. Blank lines and surrounding whitespace are removed; poems shorter than four lines use all available lines. SQL quoting is decoded by the loader. Excerpts exceeding the conservative 300-code-point or 3000-byte limit are skipped rather than cut mid-line. SQL records may contain extraction artifacts; editorial cleanup is a separate task.

Selection rotates through eligible records every 20 minutes and eventually repeats. Each 20-minute interval uses a stable record key, so repeated or concurrent invocations within that interval try to create the same record instead of creating another post. If that record already exists, the PDS rejects the create and the endpoint returns an error. Check Function Logs and the account before retrying failed or timed-out requests. No local state is used, and missed intervals are not automatically backfilled. The endpoint remains `/api/daily` for compatibility with existing deployments.

Bluesky allows 300 graphemes and 3000 bytes. Counting code points is conservative: some otherwise valid emoji or combining-character text may be excluded. See the [post schema](https://github.com/bluesky-social/atproto/blob/main/lexicons/app/bsky/feed/post.json).

## Local development

Copy `.env.example` to `.env` and fill in the values. From `dailybot`, run:

```sh
go run .
```

The server listens on port 3000 unless `PORT` is set. The protected `/api/daily` endpoint accepts GET for Vercel Cron and POST for manual invocation. An authenticated call publishes to the configured Bluesky account; visiting `/` only checks health.

Run automated tests without making real posts:

```sh
go test ./...
go vet ./...
```
