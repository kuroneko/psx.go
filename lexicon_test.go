package psx;

import (
	"testing"
)

func TestLexiconParserK(t *testing.T) {
	msg := parseMsg(nil, "Lh402(K)=KeybCduC")
	md, err := parseLexicon(msg)
	if (err != nil) {
		t.Fatalf("Failed to parse with error: %s", err)
	}
	if (md == nil) {
		t.Fatal("No message def returned")
	}
	if (md.HumanName != "KeybCduC") {
		t.Errorf("Unexpected HumanName \"%s\" received", md.HumanName)
	}
	if (md.Index != 402) {
		t.Errorf("Unexpected Index \"%d\" received", md.Index)
	}
	if (md.MessageType != MsgTypeH) {
		t.Errorf("Got wrong message type back (%d)", md.MessageType)
	}
	if (md.MessageMode != MsgModeCdukeyb) {
		t.Errorf("Got wrong message mode back (%d)", md.MessageMode)
	}
}

func TestLexiconParserZ(t *testing.T) {
	msg := parseMsg(nil, "Li242(Z)=UplinkBits")
	md, err := parseLexicon(msg)
	if (err != nil) {
		t.Fatalf("Failed to parse with error: %s", err)
	}
	if (md == nil) {
		t.Fatal("No message def returned")
	}
	if (md.HumanName != "UplinkBits") {
		t.Errorf("Unexpected HumanName \"%s\" received", md.HumanName)
	}
	if (md.Index != 242) {
		t.Errorf("Unexpected Index \"%d\" received", md.Index)
	}
	if (md.MessageType != MsgTypeI) {
		t.Errorf("Got wrong message type back (%d)", md.MessageType)
	}
	if (md.MessageMode != MsgModeXecon) {
		t.Errorf("Got wrong message mode back (%d)", md.MessageMode)
	}
}

func TestKeystring(t *testing.T) {
	msg := parseMsg(nil, "Lh402(K)=KeybCduC")
	md, _ := parseLexicon(msg)
	if md == nil {
		t.SkipNow()
	}
	keyOut := md.KeyString()
	if keyOut != "Qh402" {
		t.Errorf("Unexpected Key String: \"%s\"", keyOut)
	}
}

func TestLexicon(t *testing.T) {
	lex := newLexicon()
	err := lex.parse(parseMsg(nil, "Lh402(K)=KeybCduC"))
	if (err != nil) {
		t.Fatalf("Couldn't add Lexicon Line: %s", err)
	}
	err = lex.parse(parseMsg(nil, "Li242(Z)=UplinkBits"))
	if (err != nil) {
		t.Fatalf("Couldn't add Lexicon Line: %s", err)
	}

	msg := parseMsg(lex, "Qh402=34")
	t.Logf("Decoded Msg: %s", msg)
	t.Logf("Wire Msg: %s", msg.WireString())
	if (msg.GetDecodedKey() != "KeybCduC") {
		t.Errorf("Got unexpected key name: %s", msg.GetDecodedKey())
	}
	if (!msg.HasValue) {
		t.Error("WirePair didn't detect value in string")
		if (msg.Value != "34") {
			t.Errorf("Got unexpected value: %s", msg.Value)
		}
	}
	if (msg.WireString() != "Qh402=34") {
		t.Errorf("Reencoding got unexpected value: %s", msg.WireString())
	}
}

func TestLexiconEncode(t *testing.T) {
	lex := newLexicon()
	err := lex.parse(parseMsg(nil, "Lh402(K)=KeybCduC"))
	if (err != nil) {
		t.Fatalf("Couldn't add Lexicon Line: %s", err)
	}
	err = lex.parse(parseMsg(nil, "Li242(Z)=UplinkBits"))
	if (err != nil) {
		t.Fatalf("Couldn't add Lexicon Line: %s", err)
	}

	msg := newWireMsg(lex);
	msg.SetDecodedKey("UplinkBits");
	msg.HasValue = true
	msg.Value = "42"

	if (msg.WireString() != "Qi242=42") {
		t.Errorf("Got unexpected encoding: %s", msg.WireString())
	}
	if (msg.String() != "UplinkBits=42") {
		t.Errorf("Got unexpected display format: %s", msg)
	}
}
