package battery

//func (app *Application) rootSpawnMiddleware(next actor.SpawnFunc) actor.SpawnFunc {
//	return func(actorSystem *actor.ActorSystem, id string, props *actor.Props, parentContext actor.SpawnerContext) (*actor.PID, error) {
//		pid, err := next(actorSystem, id, props, parentContext)
//
//		app.actors.Add(pid)
//		log.Printf("rootSpawnMiddleware %v spawn %v", parentContext.Self(), pid)
//
//		return pid, err
//	}
//}
