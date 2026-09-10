# API reference

Generated from `openapi.json`. Request and response fields are described in that contract.

| Method | HTTP | Description |
| --- | --- | --- |
| `CreateSignup` | `POST /signup` | Create an account |
| `CheckSlugAvailability` | `GET /signup/check-slug` | Check if an account slug is available |
| `GetAccount` | `GET /account` | Get account stats |
| `DeleteAccount` | `DELETE /account` | Delete account permanently |
| `SendWelcomeEmail` | `POST /account/welcome` | Send the welcome email |
| `SetPostVerifyTarget` | `POST /account/post-verify-target` | Resolve post-verification target |
| `ListApiKeys` | `GET /api-keys` | List API keys |
| `CreateApiKey` | `POST /api-keys` | Create an API key |
| `RevokeApiKey` | `DELETE /api-keys/{id}` | Revoke an API key |
| `ListMailboxes` | `GET /mailboxes` | List mailboxes |
| `CreateMailbox` | `POST /mailboxes` | Create a mailbox |
| `GetMailbox` | `GET /mailboxes/{id}` | Get a mailbox |
| `UpdateMailbox` | `PATCH /mailboxes/{id}` | Update a mailbox |
| `DeleteMailbox` | `DELETE /mailboxes/{id}` | Delete a mailbox |
| `ListMessages` | `GET /mailboxes/{id}/messages` | List messages in a mailbox |
| `SendMessage` | `POST /mailboxes/{id}/messages` | Send an email from a mailbox |
| `GetMessage` | `GET /mailboxes/{id}/messages/{msgId}` | Get a message |
| `ListThreads` | `GET /mailboxes/{id}/threads` | List threads in a mailbox |
| `GetThread` | `GET /mailboxes/{id}/threads/{tid}` | Get a thread with its messages |
| `UploadAttachment` | `POST /attachments` | Upload an attachment |
| `DownloadAttachment` | `GET /attachments/{id}` | Get an attachment download URL |
| `DeleteAttachment` | `DELETE /attachments/{id}` | Delete an attachment |
| `ListDomains` | `GET /domains` | List custom domains |
| `CreateDomain` | `POST /domains` | Add a custom domain |
| `GetDomain` | `GET /domains/{id}` | Get a domain and its DNS records |
| `DeleteDomain` | `DELETE /domains/{id}` | Delete a domain |
| `VerifyDomain` | `POST /domains/{id}/verify` | Trigger domain verification |
| `ListWebhooks` | `GET /webhooks` | List webhooks |
| `CreateWebhook` | `POST /webhooks` | Create a webhook |
| `GetWebhook` | `GET /webhooks/{id}` | Get a webhook |
| `UpdateWebhook` | `PATCH /webhooks/{id}` | Update a webhook |
| `DeleteWebhook` | `DELETE /webhooks/{id}` | Delete a webhook |
| `ListWebhookDeliveries` | `GET /webhooks/{id}/deliveries` | List webhook deliveries |
| `StreamEvents` | `GET /events` | Stream inbox events over SSE |
| `ListSuppressions` | `GET /suppressions` | List suppressed addresses |
| `CreateSuppression` | `POST /suppressions` | Suppress an address |
| `DeleteSuppression` | `DELETE /suppressions/{id}` | Remove a suppression entry |
| `CreateUpgradeCheckout` | `POST /billing/upgrade` | Start a plan upgrade |
| `ResendVerificationEmail` | `POST /auth/resend-verification` | Resend the email verification link |
| `SubmitSupportTicket` | `POST /support` | Contact support |
