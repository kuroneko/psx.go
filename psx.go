// Provides an API to communicate with Precision Simulator X by Aerowinx and
// its addons.
//
// This library also supports the extensions introduced by router + switchpsx
package psx;

import (
	"net"
	"bufio"
	"strings"
	"strconv"
	"errors"
)

var (
	// Returned when actions that require a live server connection are
	// called against a closed connection
	NotConnectedError = errors.New("Connection is not currently open")
	// Returned when the PSXConn can't reconnect yet as the worker is still live.
	ConnectionBusyError = errors.New("Connection is still busy and unable to reconnect")
)

const (
	connPhaseDisconnected = iota
	connPhaseNew
	connPhaseLoad1 
	connPhaseLoad2
	connPhaseRunning
	connPhaseFailed
	connPhaseEnded
	connPhaseListenerExited
)

// MessageHooks are used for all callbacks from PSXConn's listener.
//
// The PSXConn is passed through pconn, and the message that triggered the
// callback is passed in msg.
type MessageHook func(pconn *PSXConn, msg *WireMsg)

// PSXConn manages the connection to Precision Simulator X and holds all the
// configuration.
//
// Use NewConnection to initialise a new PSXConn.
//
// Server, ClientName and InstanceName can all be changed after initialisation
// but will not be reported to the server whilst the connection is active.
//
// 
type PSXConn struct {
	// Hostname & Port to connect to 
	Server		string
	// Name of the client to report to Router/SwitchPSX
	ClientName	string
	// Name of the subinstance to report to Router/SwitchPSX
	InstanceName	string

	// Callback Hooks.
	//
	// The key is the (decoded, if necessary) attribute.
	Hooks		map[string] MessageHook

	// read-only information from the server
	myId		int 		// ID the server/router assigned us
	version		string 		// Version info as provided by the server/router
	// connection phase
	connPhase 	int 		// One of the ConnPhase* constants - defines what the current connection state is

	// notification/subscription list for SwitchPSX
	notify		[]string

	// internal bits
	conn		*net.TCPConn
	lex		*Lexicon

	bufReader 	*bufio.Reader
}

// invoke the callback with name hookName.
func (pconn *PSXConn) callHook(hookName string, msg *WireMsg) {
	callback, found := pconn.Hooks[hookName]
	if found && callback != nil {
		callback(pconn, msg)
	}
}

func NewConnection(server, myName string) (pconn *PSXConn, err error) {
	pconn = new(PSXConn)
	pconn.lex = NewLexicon()
	pconn.notify = make([]string, 0)
	pconn.connPhase = connPhaseDisconnected
	pconn.Hooks = make(map[string] MessageHook, 0)

	pconn.Server = server
	pconn.ClientName = myName

	return pconn, nil
}

// Returns the ID as assigned by the server/router
func (pconn *PSXConn) Id() int {
	return pconn.myId
}

// Returns the Software Version as reported by the server
func (pconn *PSXConn) Version() string {
	return pconn.version
}

// Connect to the server.
func (pconn *PSXConn) Connect() (err error) {
	if (nil != pconn.conn) {
		return
	}
	if (pconn.connPhase != connPhaseListenerExited && pconn.connPhase != connPhaseDisconnected) {
		return ConnectionBusyError
	}

	addr, err := net.ResolveTCPAddr("tcp", pconn.Server)
	if (err != nil) {
		return err
	}
	pconn.conn, err = net.DialTCP("tcp", nil, addr)
	if (err != nil) {
		pconn.conn = nil
		return err
	}
	pconn.connPhase = connPhaseNew
	// disable nagle explicitly - it may be the defined default, but we really want it off.
	pconn.conn.SetNoDelay(true)

	return nil
}

// Disconnect from the server.
func (pconn *PSXConn) Disconnect() {
	if (nil == pconn.conn) {
		return;
	}
	// close the reader so we can shut down propertly.
	pconn.sendLine("exit");
	pconn.conn.Close()
	pconn.conn = nil
}

// send our identity (name)
func (pconn *PSXConn) sendName() {
	nameOut := pconn.ClientName
	if (pconn.InstanceName != "") {
		nameOut += ";" + pconn.InstanceName
	}
	msgOut := pconn.NewPair("name", nameOut)
	pconn.SendMsg(msgOut)
}

// send our notify message.
func (pconn *PSXConn) sendNotify() {
	var notifyList []string = make([]string, 0)
	for _, v := range pconn.notify {
		keyName := pconn.lex.KeyFor(v)
		if keyName != "" {
			notifyList = append(notifyList, pconn.lex.KeyFor(v))
		}
	}
	if len(notifyList) > 0 {
		pconn.SendMsg(pconn.NewPair("notify", strings.Join(notifyList, ";")))
	}
}

func (pconn *PSXConn) SendMsg(msg *WireMsg) (err error) {
	return pconn.sendLine(msg.WireString())
}

func (pconn *PSXConn) sendLine(line string) (err error) {
	if nil == pconn.conn {
		return NotConnectedError
	}
	var msg []byte;

	msg = []byte(line)
	// append a CR+LF pair
	msg = append(msg, 13, 10)
	wlen, err := pconn.conn.Write(msg)
	if err != nil {
		return err
	}
	if wlen < len(msg) {
		// well crap - a short write without cause - shouldn't happen.  panic.
		panic("short write")
	}
	return nil
}


// The Listner needs to be started AFTER Connect() has been invoked.
//
// It can be started in it's own goroutine, or in the current one depending on
// requirements, but is generally intended to run in its own goroutine.
func (pconn *PSXConn) Listener() {
	var err error = nil
	running := true
	pconn.bufReader = bufio.NewReader(pconn.conn)
	for running {
		var rawLine []byte = make([]byte, 0)

		// read the full line from the network.
		var prefix bool = true
		for prefix {
			var lineSlice []byte 
			lineSlice, prefix, err = pconn.bufReader.ReadLine()
			if err != nil {
				break;
			}
			rawLine = append(rawLine, lineSlice...)
		}
		if err != nil {
			running = false
			break
		}

		// fast parse the message
		msg := ParseMsg(nil, string(rawLine))

		
		// all hard-coded reponses.
		switch (msg.GetKey()) {
		case "id":
			pconn.myId, _ = strconv.Atoi(msg.Value)
			pconn.sendName()
		case "version":
			pconn.version = msg.Value
		case "load1":
			// if we were a new connection, we were unable
			// to send notify requests until now - subscribe to our
			// desired messages.
			if (pconn.connPhase == connPhaseNew) {
				pconn.sendNotify()
			}
			pconn.connPhase = connPhaseLoad1
		case "load2":
			pconn.connPhase = connPhaseLoad2
		case "load3":
			pconn.connPhase = connPhaseRunning
		case "exit":
			pconn.connPhase = connPhaseEnded
		default:
			if (!msg.HasValue) {
				break;
			}
			if pconn.connPhase == connPhaseNew && msg.GetKey()[0] == 'L' {
				pconn.lex.Parse(msg)
			}
		}
		// once we've completed all of our integrated responses, we
		// can attempt to use the callback hooks.
		pconn.callHook(msg.GetDecodedKey(pconn.lex), msg)
	}
	if (err != nil) {
		pconn.connPhase = connPhaseFailed
	}
	pconn.Disconnect()
	pconn.connPhase = connPhaseListenerExited;
}

// Initialise a message given the human readable key/value pair
func (pconn *PSXConn) NewPair(humanKey, value string) (msg *WireMsg) {
	msg = NewWireMsg()
	msg.SetDecodedKey(pconn.lex, humanKey)
	msg.HasValue = true
	msg.Value = value
	return msg
}
