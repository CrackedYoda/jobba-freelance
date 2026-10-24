# Chat realtime protocol

The Go chat module exposes authenticated HTTP endpoints under `/api/v1/chat`, including `GET /api/v1/chat/events`, a Server-Sent Events (SSE) stream. Send the same `Authorization: Bearer <JWT>` header used by the other API endpoints. Each frame uses `message:new` or `messages:read` as its SSE event name and contains a JSON event envelope in `data`. PostgreSQL remains authoritative, so clients should refresh the affected conversation when an event arrives or after reconnecting.

This endpoint is SSE, not Socket.IO, and it does not claim Socket.IO wire compatibility. The in-process hub has bounded per-subscriber queues and global/per-user subscriber limits. Publishing never waits for a slow client; an advisory event may be dropped when that client's queue is full. Connections are removed on disconnect and receive comment keepalives. Because the hub is process-local, multi-instance deployments need sticky connections only for best-effort delivery or a shared pub/sub-backed adapter. Durable chat reads remain correct without realtime delivery.

Reverse proxies must disable response buffering for this route and allow long-lived responses. The handler sends `X-Accel-Buffering: no`, clears the Go server's normal per-response write deadline, and emits a keepalive comment every 20 seconds.

The Expo mobile app currently uses bounded foreground HTTP polling rather than the SSE endpoint. React Native does not provide a portable authenticated `EventSource` API in this dependency set. Polling pauses in the background, prevents overlapping requests, cleans up timers on unmount, and backs off after failures.
