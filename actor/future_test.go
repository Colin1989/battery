package actor

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFuture_PipeTo_Message(t *testing.T) {
	p1, mp1 := spawnMockProcess("p1")
	p2, mp2 := spawnMockProcess("p2")
	p3, mp3 := spawnMockProcess("p3")
	defer func() {
		removeMockProcess(p1)
		removeMockProcess(p2)
		removeMockProcess(p3)
	}()

	f := NewFuture(system, 1*time.Second)

	mp1.On("SendUserMessage", p1, WrapEnvelop("hello"))
	mp2.On("SendUserMessage", p2, WrapEnvelop("hello"))
	mp3.On("SendUserMessage", p3, WrapEnvelop("hello"))

	f.PipeTo(p1)
	f.PipeTo(p2)
	f.PipeTo(p3)

	ref, _ := system.ProcessRegistry.Get(f.pid)
	assert.IsType(t, &futureProcess{}, ref)
	fp, _ := ref.(*futureProcess)

	fp.SendUserMessage(f.pid, WrapEnvelop("hello"))
	mp1.AssertExpectations(t)
	mp2.AssertExpectations(t)
	mp3.AssertExpectations(t)
	assert.Empty(t, fp.pipes, "pipes were not cleared")
}

func TestFuture_PipeTo_TimeoutSendsError(t *testing.T) {
	p1, mp1 := spawnMockProcess("p1")
	p2, mp2 := spawnMockProcess("p2")
	p3, mp3 := spawnMockProcess("p3")
	defer func() {
		removeMockProcess(p1)
		removeMockProcess(p2)
		removeMockProcess(p3)
	}()

	mp1.On("SendUserMessage", p1, WrapEnvelop(ErrTimeout))
	mp2.On("SendUserMessage", p2, WrapEnvelop(ErrTimeout))
	mp3.On("SendUserMessage", p3, WrapEnvelop(ErrTimeout))

	f := NewFuture(system, 10*time.Millisecond)
	ref, _ := system.ProcessRegistry.Get(f.pid)

	f.PipeTo(p1)
	f.PipeTo(p2)
	f.PipeTo(p3)

	err := f.Wait()
	assert.Error(t, err)

	assert.IsType(t, &futureProcess{}, ref)
	fp, _ := ref.(*futureProcess)

	mp1.AssertExpectations(t)
	mp2.AssertExpectations(t)
	mp3.AssertExpectations(t)
	assert.Empty(t, fp.pipes, "pipes were not cleared")
}

type EchoResponse struct{}

func TestNewFuture_TimeoutNoRace(t *testing.T) {
	//plog.SetLevel(log.OffLevel)
	future := NewFuture(system, 1*time.Microsecond)
	a := rootContext.Spawn(PropsFromFunc(func(context Context) {
		switch context.Envelope().Message.(type) {
		case *Started:
			context.Send(future.PID(), WrapEnvelop(EchoResponse{}))
		}
	}))
	_, _ = rootContext.Request(a, WrapEnvelop(nil))
}

func assertFutureSuccess(future *Future, t *testing.T) interface{} {
	res, err := future.Result()
	assert.NoError(t, err, "timed out")
	return res
}

func TestFuture_Result_DeadLetterResponse(t *testing.T) {
	a := assert.New(t)

	//plog.SetLevel(log.OffLevel)

	future := NewFuture(system, 1*time.Second)
	rootContext.Send(future.PID(), WrapEnvelop(&DeadLetterResponse{}))
	resp, err := future.Result()
	a.Equal(ErrDeadLetter, err)
	a.Nil(resp)
}

func TestFuture_Result_Timeout(t *testing.T) {
	a := assert.New(t)

	//plog.SetLevel(log.OffLevel)

	future := NewFuture(system, 1*time.Second)
	resp, err := future.Result()
	a.Equal(ErrTimeout, err)
	a.Nil(resp)
}

func TestFuture_Result_Success(t *testing.T) {
	a := assert.New(t)

	//plog.SetLevel(log.OffLevel)

	future := NewFuture(system, 1*time.Second)
	rootContext.Send(future.PID(), WrapEnvelop(EchoResponse{}))
	resp := assertFutureSuccess(future, t)
	a.Equal(WrapEnvelop(EchoResponse{}), resp)
}
