# psx.go

A library for interfacing with Precision Simulator X from go

# License

See COPYING

# Usage

    import (
        "github.com/kuroneko/psx.go"
    )

    func doStuff() {
        conn := psx.NewConnection("localhost:10747", "myAddon")
        // set up stuff on conn
        conn.Connect()
        conn.Listener()
    }

See http://godoc.org/github.com/kuroneko/psx.go for detailed API
documentation.

It's probably important to note a few things that aren't obvious from
the API:

 * Qs/Qi/Qh strings are converted for you into their Lexicon names (see
   Variables.txt) for code readability reasons.  This translation cannot
   be trusted to occur before `load1` is received from the simulator.

 * We speak the SwitchPSX/Router extensions, including Notify.  Be sure
   to use the lexicon names, not the Q names when subscribing.

 * psx.go is threadsafe.  Just make sure you only ever start one Listener
   per PSXConn.

