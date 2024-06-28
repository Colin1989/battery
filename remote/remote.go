package remote

import "github.com/colin1989/battery/actor"

type Remote struct {
	actorSystem *actor.ActorSystem
	config      Config
}
