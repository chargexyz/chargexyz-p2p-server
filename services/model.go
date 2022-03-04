package services

// Message gets converted to/from JSON and sent in the body of pubsub messages.
type Message struct {
	Message string
	Sender  string
}
