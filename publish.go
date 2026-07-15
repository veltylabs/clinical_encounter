package clinical_encounter

// EventPublisher is implemented by the host application (SSE, websocket, etc.).
// Pass nil to disable event publishing (no-op).
type EventPublisher interface {
	Publish(event string, payload any) error
}
