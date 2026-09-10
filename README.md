# Robotomail SDK for Go

Give your application a real email address. The official Robotomail SDK covers all 40 public REST operations: mailboxes, sending and reading messages, threads, attachments, domains, API keys, webhooks, event streaming and account management.

- Runtime: **Go 1.22+**
- [API method index](API.md) · [Robotomail documentation](https://robotomail.com/docs) · [Create a mailbox](https://robotomail.com/sign-up)
- API keys authenticate these clients. For connecting an existing AI agent through OAuth, see the separate [MCP guide](https://robotomail.com/docs/mcp).

## Install

The initial release is installable from the tagged GitHub repository:

```sh
go get github.com/robotomail/robotomail-go@v0.1.0
```

Go modules use this repository and its version tags directly.

## Quickstart

Set `ROBOTOMAIL_API_KEY` in your application's environment. Keep keys on the server. Use a mailbox-scoped key when an application only needs selected mailboxes; account operations and attachment uploads require a full-access key.

```go
package main

import (
    "context"
    "fmt"
    "log"
    robotomail "github.com/robotomail/robotomail-go"
)

func main() {
    mail, err := robotomail.NewClient(robotomail.Options{}) // reads ROBOTOMAIL_API_KEY
    if err != nil { log.Fatal(err) }
    result, err := mail.ListMailboxes(context.Background())
    if err != nil { log.Fatal(err) }
    for _, box := range result.Mailboxes { fmt.Println(box.FullAddress) }
}
```

The runnable example in `examples/list-mailboxes/main.go` lists mailboxes and does not send email. Create an account and verify your email first. The Free plan includes one mailbox and 10 sends plus 10 receives per calendar month, restricted to your verified email address. Upgrade when you need more.

## Set a sender name and send

Use an owned mailbox ID and replace the example recipient before running:

```go
_, err := mail.UpdateMailbox(ctx, mailboxID, robotomail.UpdateMailboxRequest{
    DisplayName: robotomail.Ptr("Research Agent"),
})
if err != nil { return err }
_, err = mail.SendMessage(ctx, mailboxID, robotomail.SendMessageRequest{
    To: []string{"you@example.com"}, Subject: "Research complete",
    BodyText: "Here are my findings.",
})
if err != nil { return err }
```

Messages use `inReplyTo` to thread a reply. The API returns the sent message after the outbound provider accepts it; that does not mean the recipient has received it yet.

## Read and paginate

Use the list-messages method with `limit` (1–100, default 50) and `offset` (default 0). Increase the offset by the number returned and stop when a page contains fewer than the requested limit. Filters include `direction`, `threadId` and `since`. Lists are newest first and can change while paging, so retain message IDs to deduplicate long-running scans. Over-quota withheld messages are reported in metadata and excluded from the returned list.

## Live events

Streaming returns frames with `id`, `event` and raw `data`. Parse `data` as JSON for email events. Save the last event ID when one is supplied.

```go
var lastEventID string
events, err := mail.StreamEvents(ctx, &robotomail.StreamEventsParams{
    MailboxId: robotomail.Ptr(mailboxID),
})
if err != nil { return err }
defer events.Close()
for events.Next() {
    frame := events.Event()
    lastEventID = frame.ID
    if frame.Event == "reconnect" { break }
    fmt.Println(frame.Event, frame.Data)
}
if err := events.Err(); err != nil { return err }
// Reopen with LastEventID: robotomail.Ptr(lastEventID) to resume.
```

Each stream is one connection. The server asks clients to reconnect after about 4.5 minutes. Reopen with `Last-Event-ID`; replay is limited to the last 100 events within one hour. Apply your own backoff and cancellation policy. A stream alone does not schedule an agent to process mail.

## Attachments

The upload method accepts binary data, a filename and a content type, using multipart field `file`. The maximum attachment size is 25 MB. Sending uses the returned attachment ID in `attachments`. The download-attachment operation returns attachment metadata and a presigned `url`; fetch that URL separately without adding your Robotomail API key.

## Webhook verification

Use `VerifyWebhook` with the original request body, `X-Robotomail-Signature`, and the webhook secret. It checks the HMAC-SHA256 signature in constant time. Pass the raw bytes before JSON parsing; whitespace changes or reserializing JSON invalidate the signature. The signature does not contain a timestamp, so deduplicate events if your handler needs replay protection.

## Errors, timeouts and retries

HTTP failures expose status, the original error body and response headers. This includes permission errors, Free-plan restrictions, rate limits and upstream non-JSON failures. Quota errors may include a machine-readable `code` and `Retry-After` or reset information. Transport failures and cancellation remain distinct from HTTP errors.

Normal requests default to a 30-second timeout. Streams allow longer connections. The SDK never automatically repeats send operations: a timeout may happen after an email was accepted, so inspect the mailbox before retrying. Redirects are not followed. Override the API base URL for testing with a full `/v1` URL; HTTPS is required except on loopback development servers.

Optional request fields are omitted unless supplied; explicit null is preserved. Response types accept unknown enum strings for forward compatibility. The server remains responsible for validating requests. Regenerate types to expose newly introduced response fields.

## Development

```sh
go test -race ./...
go vet ./...
python3.12 scripts/generate.py --check
```

Tests use a loopback HTTP fixture with fake credentials. They exercise all 40 operations, URL/query encoding, request and response bodies, binary uploads, streaming, error headers, timeouts, redirects and webhook verification. No production account or real email is needed.

`openapi.json` is the pinned public contract. After reviewing an updated contract, run `python3.12 scripts/generate.py`; it replaces only models, operation bindings, `API.md` and `contract.json`. Hand-maintained transport code is preserved. Go must be installed for formatting. Contract fixtures are explicit test inputs; review and extend them when adding API operations.

See [CONTRIBUTING.md](CONTRIBUTING.md) for releases. Licensed under [MIT](LICENSE).

## Optional and nullable Go fields

Use `robotomail.Ptr(value)` for optional fields. For example, clear webhook headers with `UpdateWebhookRequest{Headers: robotomail.Ptr[*robotomail.WebhookHeaders](nil)}`. Leaving `Headers` nil omits the field; a non-nil outer pointer containing nil sends explicit JSON null. Dates remain ISO-8601 strings as returned by the API.
