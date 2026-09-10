// Run only with the application's disposable local fixture.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	rm "github.com/robotomail/robotomail-go"
	"net/url"
	"os"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func assert(ok bool) {
	if !ok {
		panic("SDK integration assertion failed")
	}
}
func main() {
	var f struct{ BaseURL, APIKey, ScopedKey, UserID, MailboxID, ForeignMailboxID, MessageID, Body string }
	check(json.Unmarshal([]byte(os.Getenv("ROBOTOMAIL_SDK_TEST")), &f))
	u, err := url.Parse(f.BaseURL)
	check(err)
	assert(u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1")
	full, err := rm.NewClient(rm.Options{APIKey: f.APIKey, BaseURL: f.BaseURL})
	check(err)
	scoped, err := rm.NewClient(rm.Options{APIKey: f.ScopedKey, BaseURL: f.BaseURL})
	check(err)
	ctx := context.Background()
	account, err := full.GetAccount(ctx)
	check(err)
	assert(account.Account.Id == f.UserID)
	boxes, err := full.ListMailboxes(ctx)
	check(err)
	assert(boxes.Mailboxes[0].Id == f.MailboxID)
	messages, err := scoped.ListMessages(ctx, f.MailboxID, &rm.ListMessagesParams{Limit: rm.Ptr[int64](1), Offset: rm.Ptr[int64](0)})
	check(err)
	assert(messages.Messages[0].Id == f.MessageID)
	message, err := full.GetMessage(ctx, f.MailboxID, f.MessageID)
	check(err)
	assert(message.Message.BodyText == f.Body)
	_, err = scoped.UpdateMailbox(ctx, f.MailboxID, rm.UpdateMailboxRequest{DisplayName: rm.Ptr("SDK integration")})
	check(err)
	box, err := full.GetMailbox(ctx, f.MailboxID)
	check(err)
	assert(*box.Mailbox.DisplayName == "SDK integration")
	_, err = scoped.GetAccount(ctx)
	var api *rm.APIError
	assert(errors.As(err, &api) && api.Status == 403)
	_, err = scoped.GetMailbox(ctx, f.ForeignMailboxID)
	assert(errors.As(err, &api) && api.Status == 404)
	hook, err := scoped.CreateWebhook(ctx, rm.CreateWebhookRequest{Url: "https://example.com/sdk-fixture", MailboxId: rm.Ptr(f.MailboxID), Events: []string{"message.received"}})
	check(err)
	_, err = full.UpdateWebhook(ctx, hook.Webhook.Id, rm.UpdateWebhookRequest{Headers: rm.Ptr[*rm.WebhookHeaders](nil)})
	check(err)
	got, err := full.GetWebhook(ctx, hook.Webhook.Id)
	check(err)
	assert(got.Webhook.Headers == nil)
	_, err = full.DeleteWebhook(ctx, hook.Webhook.Id)
	check(err)
	sent, err := full.SendMessage(ctx, f.MailboxID, rm.SendMessageRequest{To: []string{"delivered@resend.dev"}, Subject: "Local SDK test", BodyText: "No external email is delivered."})
	check(err)
	assert(sent.Message.Status == "SENT" && strings.HasPrefix(*sent.Message.ExternalMessageId, "email-mock-"))
	gotMessage, err := scoped.GetMessage(ctx, f.MailboxID, sent.Message.Id)
	check(err)
	assert(gotMessage.Message.Subject == "Local SDK test")
	result, _ := json.Marshal(map[string]any{"checks": 12, "sentMessageId": sent.Message.Id})
	fmt.Println(string(result))
}
