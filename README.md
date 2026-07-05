# workflow-plugin-messaging-core

> ⚠️ **Experimental** — This plugin compiles and passes its unit tests but has not been validated in any active GoCodeAlone-internal production deployment. Use with caution. Please [open an issue](https://github.com/GoCodeAlone/workflow-plugin-messaging-core/issues/new) if you adopt it so we can promote it to **verified** status.

Shared messaging interfaces and step contracts for workflow platform messaging plugins (Discord, Slack, Teams).

## Install

```sh
wfctl plugin install workflow-plugin-messaging-core
```

## Ratchet notification handoff

`ratchet blackboard export [section] --jsonl` emits local notification-event
records with `messaging.text`. This package can parse those JSON or JSONL
records with `ParseRatchetNotificationEvents` and project them into typed
`step.messaging_send` input with `ProjectRatchetNotificationToMessagingSend`.

Ratchet does not send Slack, Discord, Teams, webhook, or email messages
directly. Downstream Workflow pipelines provide the target `channel` and use the
platform plugins that own credentials, rate limits, redaction, retries, and
delivery.

## License

MIT
