package cb

import (
	"fmt"
	"testing"
	"time"
)

type testConsumer struct {
	ok    bool
	delay time.Duration
}

func (c *testConsumer) init() {
	for {
		<-time.After(c.delay)
		if c.ok {
			fmt.Println("CLOSED CONSUMER")
			c.ok = false
			continue
		}

		fmt.Println("OPENED CONSUMER")
		c.ok = true
	}
}

func (c *testConsumer) exec() error {
	if !c.ok {
		return err("consumer is not responding")
	}
	fmt.Println("OK!")
	<-time.After(100 * time.Millisecond)
	return nil
}

func TestCircuitBreaker_Exec(t *testing.T) {
	consumer := &testConsumer{false, 500 * time.Millisecond}
	go consumer.init()

	cb := NewBufferedBreaker(2, 3, 300*time.Millisecond)
	res := cb.Results()
	for i := 0; i < 10; i++ {
		t.Log("sending #", i)
		if err := cb.Exec(consumer.exec); err != nil {
			t.Log(err)
		}
	}
	lim := 10

	for {
		select {
		case <-res:
			lim--
			if lim == 0 {
				t.Log("retrieved 10 records as expected")
				return
			}
		case <-time.After(time.Second * 5):
			t.Error("timeout")
		}
	}

}
