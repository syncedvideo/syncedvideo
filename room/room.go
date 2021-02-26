package room

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID    fmt.Stringer `json:"id" db:"id"`
	users map[User]bool
	join  chan User
	leave chan User
}

type User interface {
	ID() fmt.Stringer
}

func New() Room {
	return Room{
		ID:    uuid.New(),
		users: make(map[User]bool),
		join:  make(chan User),
		leave: make(chan User),
	}
}

func (r *Room) Run() {
	for {
		time.Sleep(time.Millisecond)
		select {
		case user := <-r.join:
			r.users[user] = true
		case user := <-r.leave:
			delete(r.users, user)
		}
	}
}

func (r *Room) Close() {
	close(r.join)
}
