package psx

import (
	"strings"
	"strconv"
	"errors"
)

var (
	DuplicateNameError = errors.New("lexicon Name Already Registered")
	DuplicateIndexError = errors.New("lexicon Index Already Registered")
	lexiconSyntaxError = errors.New("lexicon Line Has Invalid Syntax")
	NotAQStringError = errors.New("Line cannot be decoded - not a Q string")
	QStringSyntaxError = errors.New("Malformed Q String")
	UnknownIndexError = errors.New("Index not known to lexicon")
)

const (
	MsgTypeI = iota 		// Integer Parameter (Qi)
	MsgTypeS 			// String Parameter (Qs)
	MsgTypeH 			// Human Paramter (Qh)
)

const (
	MsgModeStart = iota 		// (S)
	MsgModeCont 			// (C) ? never seen
	MsgModeEcon 			// (E)
	MsgModeDelta 			// (D)
	MsgModeBigmom 			// (B)
	MsgModeMcpmom 			// (M)
	MsgModeGuamom2 			// (G)
	MsgModeGuamom4 			// (F)
	MsgModeCdukeyb 			// (K)
	MsgModeRcp 			// (R)
	MsgModeAcp			// (A)
	MsgModeMixed 			// (X)
	MsgModeXdelta 			// (Y)
	MsgModeXecon			// (Z)
	MsgModeDemand			// (N)
)

// MessageDef defines a single term in the Precision Simulator lexicon.
//
// A MessageDef contains the information you need to map the name to and from
// it's number (Index) as well as the information given to us about its
// update mode.
type MessageDef struct {
	MessageType	int	// one of the MsgType constants reflecting the RHS format.
	MessageMode	int 	// onde of the MsgMode constants reflecting the update frequency/type
	Index		int	// numeric index within the given MessageType for the message
	HumanName 	string	// the humanish display name for the item
}

func (msgdef *MessageDef) KeyString() string {
	switch (msgdef.MessageType) {
	case MsgTypeI:
		return "Qi" + strconv.Itoa(msgdef.Index)
	case MsgTypeS:
		return "Qs" + strconv.Itoa(msgdef.Index)
	case MsgTypeH:
		return "Qh" + strconv.Itoa(msgdef.Index)
	}
	// dunno how to handle this - return empty.
	return ""
}

// parse a raw lexicon Line from a server into a defintion.
func parseLexicon(lexMsg *WireMsg) (msgdef *MessageDef, err error) {
	msgdef = new(MessageDef)
	key := lexMsg.GetKey()
	if key[0] != 'L' {
		return nil, lexiconSyntaxError
	}
	// must be at least 6 charts long for the L + type (2 chars), the index (1 char min) and mode suffix (3 chars).
	if len(key) < 6 {
		return nil, lexiconSyntaxError
	}
	// parse type.
	// naive parse should be enough.
	switch (key[1]) {
	case 'i':
		msgdef.MessageType = MsgTypeI
	case 's':
		msgdef.MessageType = MsgTypeS
	case 'h':
		msgdef.MessageType = MsgTypeH
	default:
		return nil, lexiconSyntaxError
	}
	// now, split out the number
	suffixIdx := strings.Index(key, "(")
	if (suffixIdx < 0) {
		return nil, lexiconSyntaxError
	}
	msgdef.Index, err = strconv.Atoi(key[2:suffixIdx])
	if (err != nil) {
		return nil, err
	}
	// and the type
	switch (key[suffixIdx+1]) {
	case 'S':
		msgdef.MessageMode = MsgModeStart
	case 'C':
		msgdef.MessageMode = MsgModeCont
	case 'E':
		msgdef.MessageMode = MsgModeEcon
	case 'D':
		msgdef.MessageMode = MsgModeDelta
	case 'B':
		msgdef.MessageMode = MsgModeBigmom
	case 'M':
		msgdef.MessageMode = MsgModeMcpmom
	case 'G':
		msgdef.MessageMode = MsgModeGuamom2
	case 'F':
		msgdef.MessageMode = MsgModeGuamom4
	case 'K':
		msgdef.MessageMode = MsgModeCdukeyb
	case 'R':
		msgdef.MessageMode = MsgModeRcp
	case 'A':
		msgdef.MessageMode = MsgModeAcp
	case 'X':
		msgdef.MessageMode = MsgModeMixed
	case 'Y':
		msgdef.MessageMode = MsgModeXdelta
	case 'Z':
		msgdef.MessageMode = MsgModeXecon
	case 'N':
		msgdef.MessageMode = MsgModeDemand
	default:
		return nil, lexiconSyntaxError
	}
	msgdef.HumanName = lexMsg.Value
	return msgdef, nil
}

// A lexicon holds the data necessary to dynamically learn and map the PSX 
// lexicon for Qi/Qh/Qs messages so we can use the human names internally
//
// This allows for (hopefully) less painful to read code.
type lexicon struct {
	forward 	map[string] *MessageDef	// forward lookup stores the Qh/Qs/Qi to messagedef map
	reverse		map[string] *MessageDef // reverse lookup stores the humanName to Qh/Qs/Qi map
}

// initialise a new, empty, lexicon ready to be filled with mappings
func newLexicon() (lex *lexicon) {
	lex = new(lexicon);
	lex.forward = make(map[string] *MessageDef, 0)
	lex.reverse = make(map[string] *MessageDef, 0)

	return lex
}

// Finds the Q key for a given named paramater.  returns the empty string
// if it can't find it.
func (lex *lexicon) keyFor(humanName string) string {
	def, found := lex.reverse[humanName]
	if (found) {
		return def.KeyString()
	} else {
		return ""
	}
}

// given a Qstring, find the human name.  returns the empty string if it can't
// find the mapping.
func (lex *lexicon) humanNameFor(keyName string) string {
	def, found := lex.forward[keyName]
	if (found) {
		return def.HumanName
	} else {
		return ""
	}	
}

func (lex *lexicon) parse(msgIn *WireMsg) (err error) {
	md, err := parseLexicon(msgIn)
	if (err != nil) {
		return err
	}
	lex.reverse[md.HumanName] = md
	lex.forward[md.KeyString()] = md

	return nil
}
