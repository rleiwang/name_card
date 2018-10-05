package main

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type broker struct {
	// Registered clients.
	clients map[*client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *client

	// Unregister requests from clients.
	unregister chan *client

	control chan bool
}

func newBroker() *broker {
	return &broker{
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
		clients:    make(map[*client]bool),
		control:    make(chan bool),
	}
}

func (b *broker) run() {
	for {
		select {
		case client := <-b.register:
			b.clients[client] = true
		case client := <-b.unregister:
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.send)
			}
		case message := <-b.broadcast:
			for client := range b.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(b.clients, client)
				}
			}
		case <-b.control:
			return
		}
	}
}

func (b *broker) stop() {
	close(b.control)
}
