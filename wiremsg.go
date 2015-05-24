package psx

import (
	"strings"
	"fmt"
)

// Represents a single message off the wire from the server
//
// Construct an empty WireMsg using NewWireMsg() or construct it from a line
// of protocol using ParseMsg().
type WireMsg struct {
	// Encoded (Wire) Key/Action (left hand side)
  	key   	string
  	// Indicates if there's a data section (right hand side)
	HasValue	bool
	// Value of the data section (right hand side)
	Value		string

	// cached message defintion for this WireMsg
	definition	*MessageDef
}

// Initialise a new (blank) WireMsg
func NewWireMsg() (msg *WireMsg) {
	msg = new(WireMsg)

	return msg
}

// Parse a line of input from the server and return it in WireMsg form
func ParseMsg(lex *Lexicon, line string) (msg *WireMsg) {
	msg = NewWireMsg()
	msg.Parse(lex, line)
	return msg

}

// relink the definition against the key (or clear it so the next attempt can
//    retry it)
func (msg *WireMsg) relinkKey(lex *Lexicon) {
	if lex != nil {
		msg.definition, _ = lex.forward[msg.key]
	} else {
		msg.definition = nil
	}
}

// Populate this WireMsg with the line of network input (sans line end) 
func (msg *WireMsg) Parse(lex *Lexicon, line string) {
	if strings.Index(line, "=") < 0 {
		msg.HasValue = false
		msg.SetKey(line)
	} else {
		parts := strings.SplitN(line, "=", 2)
		msg.SetKey(parts[0])
		msg.HasValue = true
		msg.Value = parts[1]
	}
	// relink using the lexicon
	msg.relinkKey(lex)
}

// try to decode the message key using the lexicon (using any
// cached result).  Return the key if there's no decoded value.
//
// Will cache the result if none exists, so you can use this (and discard
// the value) to force a late decode
func (msg *WireMsg) GetDecodedKey(lex *Lexicon) (string) {
	if (msg.definition == nil && lex != nil) {
		msg.relinkKey(lex)
	}
	if msg.definition != nil {
		return msg.definition.HumanName
	}
	return msg.key
}

// given the humanName, set the Key.
func (msg *WireMsg) SetDecodedKey(lex *Lexicon, humanName string) {
	var found = false
	var def *MessageDef = nil

	if lex != nil {
		def, found = lex.reverse[humanName]
		if found {
			msg.definition = def
			msg.key = def.KeyString()
		}
	}
	if !found {
		msg.definition = nil
		msg.key = humanName
	}
}

// set the key without any decode attempt
func (msg *WireMsg) SetKey(key string) {
	if (msg.key != key) {
		defer msg.relinkKey(nil);
	}
	msg.key = key
}

// return the encoded key as it would be sent on the wire
func (msg *WireMsg) GetKey() string {
	return msg.key
}

// return the message, ready to send, as a string.
func (msg *WireMsg) WireString() string {
	if (msg.HasValue) {
		return fmt.Sprintf("%s=%s", msg.key, msg.Value)
	} else {
		return msg.key
	}
}

// string format the message with the key decoded for easy logging/debug use.
func (msg *WireMsg) String() string {
	if (msg.HasValue) {
		return fmt.Sprintf("%s=%s", msg.GetDecodedKey(nil), msg.Value)
	} else {
		return msg.key
	}
}

// asumming the value is ; delimited, get the value at the numbered subindex.
// found will be true if it was there, false otherwise.
func (msg *WireMsg) ValueAtSubIndex(idx int) (val string, found bool) {
	if !msg.HasValue {
		return "", false
	}
	parts := strings.Split(msg.Value, ";")
	if (idx >= len(parts)) {
		return "", false
	}
	return parts[idx], true
}
