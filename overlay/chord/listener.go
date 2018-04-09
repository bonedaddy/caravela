package chord

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bluele/go-chord"
	nodeAPI "github.com/strabox/caravela/node/api"
	"math/big"
)

type Listener struct {
	thisNode nodeAPI.OverlayMembership // Caravela Node on top of the chord overlay
}

func (cl *Listener) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	if local != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(local.Id)
		cl.thisNode.AddTrader(local.Id)
		log.Debugf("[Chord] Local Node: ID: %s IP: %s", idToPrint.String(), local.Host)
	}
	if remoteNew != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(remoteNew.Id)
		log.Debugf("[Chord] Remote Node: ID: %s IP: %s", idToPrint.String(), remoteNew.Host)
	}
	if remotePrev != nil {
		idToPrint := big.NewInt(0)
		idToPrint.SetBytes(remotePrev.Id)
		log.Debugf("[Chord] Previous Remote Node: ID: %s IP: %s", idToPrint.String(), remotePrev.Host)
	}
}

func (cl *Listener) Leaving(local, predecessor, successor *chord.Vnode) {
	log.Debugln("[Chord] I am leaving!!")
}

func (cl *Listener) PredecessorLeaving(local, remote *chord.Vnode) {
	log.Debugln("[Chord] Current predecessor is leaving!!")
}

func (cl *Listener) SuccessorLeaving(local, remote *chord.Vnode) {
	log.Debugln("[Chord] A successor is leaving!!")
}

func (cl *Listener) Shutdown() {
	log.Debugln("[Chord] Shutting Down!!")
}