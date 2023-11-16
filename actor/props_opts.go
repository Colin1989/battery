package actor

type PropsOption func(props *Props)

// PropsFromProducer creates a props with the given actor producer assigned.
func PropsFromProducer(producer Producer, opts ...PropsOption) *Props {
	p := &Props{
		producer: producer,
	}
	p.Configure(opts...)

	return p
}

// PropsFromFunc creates a props with the given receive func assigned as the actor producer.
func PropsFromFunc(f ReceiveFunc, opts ...PropsOption) *Props {
	p := PropsFromProducer(func() Actor { return f }, opts...)

	return p
}
