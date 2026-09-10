// Contract call adapter. Assertions and HTTP fixture are maintained separately.
package robotomail

import (
	"context"
	"encoding/json"
	"fmt"
)

func callCase(c *Client, name string, rawBody, rawParams json.RawMessage) (any, error) {
	ctx := context.Background()
	switch name {
	case "createSignup":
		var body SignupRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.CreateSignup(ctx, body)
	case "checkSlugAvailability":
		var params CheckSlugAvailabilityParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, err
		}
		return c.CheckSlugAvailability(ctx, &params)
	case "getAccount":
		return c.GetAccount(ctx)
	case "deleteAccount":
		var body DeleteAccountRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.DeleteAccount(ctx, body)
	case "sendWelcomeEmail":
		return c.SendWelcomeEmail(ctx)
	case "setPostVerifyTarget":
		return c.SetPostVerifyTarget(ctx)
	case "listApiKeys":
		return c.ListApiKeys(ctx)
	case "createApiKey":
		var body CreateApiKeyRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.CreateApiKey(ctx, body)
	case "revokeApiKey":
		return c.RevokeApiKey(ctx, "box /?#%")
	case "listMailboxes":
		return c.ListMailboxes(ctx)
	case "createMailbox":
		var body CreateMailboxRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.CreateMailbox(ctx, body)
	case "getMailbox":
		return c.GetMailbox(ctx, "box /?#%")
	case "updateMailbox":
		var body UpdateMailboxRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.UpdateMailbox(ctx, "box /?#%", body)
	case "deleteMailbox":
		return c.DeleteMailbox(ctx, "box /?#%")
	case "listMessages":
		var params ListMessagesParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, err
		}
		return c.ListMessages(ctx, "box /?#%", &params)
	case "sendMessage":
		var body SendMessageRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.SendMessage(ctx, "box /?#%", body)
	case "getMessage":
		return c.GetMessage(ctx, "box /?#%", "child /?#%")
	case "listThreads":
		return c.ListThreads(ctx, "box /?#%")
	case "getThread":
		return c.GetThread(ctx, "box /?#%", "child /?#%")
	case "uploadAttachment":
		return c.UploadAttachment(ctx, Upload{Data: []byte{0, 255, 98, 105, 110, 97, 114, 121, 13, 10}, Filename: "sample.bin"})
	case "downloadAttachment":
		return c.DownloadAttachment(ctx, "box /?#%")
	case "deleteAttachment":
		return c.DeleteAttachment(ctx, "box /?#%")
	case "listDomains":
		return c.ListDomains(ctx)
	case "createDomain":
		var body CreateDomainRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.CreateDomain(ctx, body)
	case "getDomain":
		return c.GetDomain(ctx, "box /?#%")
	case "deleteDomain":
		return c.DeleteDomain(ctx, "box /?#%")
	case "verifyDomain":
		return c.VerifyDomain(ctx, "box /?#%")
	case "listWebhooks":
		return c.ListWebhooks(ctx)
	case "createWebhook":
		var body CreateWebhookRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.CreateWebhook(ctx, body)
	case "getWebhook":
		return c.GetWebhook(ctx, "box /?#%")
	case "updateWebhook":
		var body UpdateWebhookRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		body.Headers = Ptr[*WebhookHeaders](nil)
		return c.UpdateWebhook(ctx, "box /?#%", body)
	case "deleteWebhook":
		return c.DeleteWebhook(ctx, "box /?#%")
	case "listWebhookDeliveries":
		return c.ListWebhookDeliveries(ctx, "box /?#%")
	case "streamEvents":
		var params StreamEventsParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, err
		}
		return c.StreamEvents(ctx, &params)
	case "listSuppressions":
		return c.ListSuppressions(ctx)
	case "createSuppression":
		var body CreateSuppressionRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.CreateSuppression(ctx, body)
	case "deleteSuppression":
		return c.DeleteSuppression(ctx, "box /?#%")
	case "createUpgradeCheckout":
		var body UpgradeRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.CreateUpgradeCheckout(ctx, body)
	case "resendVerificationEmail":
		var body ResendVerificationEmailRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.ResendVerificationEmail(ctx, body)
	case "submitSupportTicket":
		var body SupportRequest
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return nil, err
		}
		return c.SubmitSupportTicket(ctx, body)
	default:
		return nil, fmt.Errorf("unknown operation %s", name)
	}
}
