package godave

import (
	"github.com/disgoorg/snowflake/v2"

	"github.com/disgoorg/disgo/voice"
)

type Dave struct{}

func NewDave() *Dave {
	return &Dave{}
}

func (d *Dave) CreateSession(userID snowflake.ID, callbacks voice.Callbacks) voice.DaveSession {
	return NewSession(userID, callbacks)
}
