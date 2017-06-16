package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/muniverse"
	"github.com/unixpickle/muniverse/chrome"
)

// An EnvHandler handles a WebSocket for the remote
// environment API.
type EnvHandler struct {
	Server *Server
	Conn   *websocket.Conn
	Spec   *muniverse.EnvSpec

	env  muniverse.Env
	done bool
}

// Handle reads and executes commands from the connection
// until the first unrecoverable error occurs.
func (e *EnvHandler) Handle() error {
	defer func() {
		if e.env != nil {
			e.env.Close()
		}
	}()
	for {
		var msg *envMessage
		if err := e.Conn.ReadJSON(&msg); err != nil {
			return err
		}
		var err error
		switch msg.Type {
		case resetMessage:
			err = e.reset()
		case stepMessage:
			err = e.step(msg)
		default:
			err = fmt.Errorf("invalid message type: %s", msg.Type)
		}
		if err != nil {
			e.sendError(err)
			return err
		}
	}
}

func (e *EnvHandler) reset() (err error) {
	defer essentials.AddCtxTo("reset", &err)
	if e.env == nil {
		e.env, err = muniverse.NewEnv(e.Spec)
		if err != nil {
			return err
		}
	}
	if err = e.env.Reset(); err != nil {
		return err
	}

	msg, err := e.messageWithObs(resetMessage)
	if err != nil {
		return err
	}
	return e.Conn.WriteJSON(msg)
}

func (e *EnvHandler) step(msg *envMessage) (err error) {
	defer essentials.AddCtxTo("step", &err)
	if e.env == nil || e.done {
		return errors.New("cannot step without a reset")
	}
	var events []interface{}
	for _, evt := range msg.Actions {
		if evt.KeyEvent != nil {
			events = append(events, evt.KeyEvent)
		} else if evt.MouseEvent != nil {
			events = append(events, evt.MouseEvent)
		}
	}
	reward, done, err := e.env.Step(e.Server.FrameTime, events)
	if err != nil {
		return err
	}
	if done {
		e.done = true
	}
	resMsg, err := e.messageWithObs(stepMessage)
	if err != nil {
		return err
	}
	resMsg.Reward = reward
	resMsg.Done = done
	return e.Conn.WriteJSON(resMsg)
}

func (e *EnvHandler) sendError(err error) error {
	msg := err.Error()
	return e.Conn.WriteJSON(&envMessage{
		Type:  errorMessage,
		Error: &msg,
	})
}

func (e *EnvHandler) messageWithObs(t envMessageType) (*envMessage, error) {
	obs, err := e.env.Observe()
	if err != nil {
		return nil, err
	}

	data, err := muniverse.ObsPNG(obs)
	if err != nil {
		return nil, err
	}
	var b64 bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &b64)
	enc.Write(data)
	enc.Close()
	b64Str := b64.String()

	return &envMessage{
		Type:        t,
		Observation: &b64Str,
	}, nil
}

// envMessageType is a message type ID used to distinguish
// between types of messages in the environment API.
type envMessageType string

const (
	resetMessage envMessageType = "reset"
	stepMessage                 = "step"
	errorMessage                = "error"
)

// envMessage is a message object for the environment API.
type envMessage struct {
	Type envMessageType `json:"type"`

	// Used for Reset and Step responses.
	Observation *string `json:"observation,omitempty"`

	// Used for Step responses.
	Reward float64 `json:"reward"`
	Done   bool    `json:"done"`

	// Used for Step requests.
	Actions []*envAction `json:"actions,omitempty"`

	// Used for Error responses.
	Error *string `json:"error,omitempty"`
}

// An envAction is an action in the environment API.
type envAction struct {
	KeyEvent   *chrome.KeyEvent   `json:"keyEvent"`
	MouseEvent *chrome.MouseEvent `json:"mouseEvent"`
}
