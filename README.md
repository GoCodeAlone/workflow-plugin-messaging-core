# workflow-plugin-messaging-core

> ⚠️ **Experimental** — This plugin compiles and passes its unit tests but has not been validated in any active GoCodeAlone-internal production deployment. Use with caution. Please [open an issue](https://github.com/GoCodeAlone/workflow-plugin-messaging-core/issues/new) if you adopt it so we can promote it to **verified** status.

Shared messaging interfaces and step contracts for workflow platform messaging plugins (Discord, Slack, Teams).

## Install

```sh
wfctl plugin install workflow-plugin-messaging-core
```

## Notification-event handoff

Producers can hand off local status, coordination, or release messages as JSON
or JSONL notification-event records. Each record carries `messaging.text`, and
may include Workflow handoff metadata for `step.messaging_send`.

Use `ParseNotificationEvents` to parse a JSON array or JSONL stream and
`ProjectNotificationEventToMessagingSend` to project a record into typed
`step.messaging_send` input. The downstream Workflow pipeline supplies the
target `channel`; provider-specific plugins such as Slack, Discord, or Teams
own credentials, rate limits, redaction, retries, and delivery.

Ratchet blackboard export is one producer of this shape. Deprecated
ratchet-prefixed aliases remain for compatibility, but new callers should use
the generic notification-event API.

## License

MIT
