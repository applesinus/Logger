package constants

import (
	"errors"

	"github.com/rs/zerolog"
)

var (
	ErrInvalidLevel = errors.New("LOGGER Invalid level")
	ErrLevelTooLow  = errors.New("LOGGER Level of msg is lower than cap level")
	ErrBufferIsFull = errors.New("LOGGER Buffer is full")
)

var MtzlLevel = zerolog.Level(6)
