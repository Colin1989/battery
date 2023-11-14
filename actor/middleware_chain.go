package actor

func makeSenderMiddlewareChain(senderMiddleware []SenderMiddleware, lastSender SenderFunc) SenderFunc {
	if len(senderMiddleware) == 0 {
		return nil
	}

	h := senderMiddleware[len(senderMiddleware)-1](lastSender)
	for i := len(senderMiddleware) - 2; i >= 0; i-- {
		h = senderMiddleware[i](h)
	}

	return h
}

func makeSpawnMiddlewareChain(spawnMiddleware []SpawnMiddleware, lastSpawn SpawnFunc) SpawnFunc {
	if len(spawnMiddleware) == 0 {
		return nil
	}

	h := spawnMiddleware[len(spawnMiddleware)-1](lastSpawn)
	for i := len(spawnMiddleware) - 2; i >= 0; i-- {
		h = spawnMiddleware[i](h)
	}

	return h
}
