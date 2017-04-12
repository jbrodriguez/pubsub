// Copyright 2013, Chandra Sekar S.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the README.md file.
// Copyright 2017, Juan B. Rodriguez
// The modifications to this package made by
// Juan B. Rodriguez, are governed by a MIT license,
// that can be found in the LICENSE file.

package pubsub

import (
	"runtime"
	"testing"
	"time"

	"gopkg.in/check.v1"
)

var _ = check.Suite(new(Suite))

func Test(t *testing.T) {
	check.TestingT(t)
}

type Suite struct{}

func (s *Suite) TestSub(c *check.C) {
	ps := New(1)
	mb1 := ps.CreateMailbox()
	mb2 := ps.CreateMailbox()
	mb3 := ps.CreateMailbox()
	ps.Sub(mb1, "t1")
	ps.Sub(mb2, "t1")
	ps.Sub(mb3, "t2")

	ps.Pub(&Message{Payload: "hi"}, "t1")
	c.Check(<-mb1, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})
	c.Check(<-mb2, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})

	ps.Pub(&Message{Payload: "hello"}, "t2")
	c.Check(<-mb3, check.DeepEquals, &Mailbox{Topic: "t2", Content: &Message{Payload: "hello"}})

	ps.Shutdown()
	_, ok := <-mb1
	c.Check(ok, check.Equals, false)
	_, ok = <-mb2
	c.Check(ok, check.Equals, false)
	_, ok = <-mb3
	c.Check(ok, check.Equals, false)
}

func (s *Suite) TestSubOnce(c *check.C) {
	ps := New(1)
	mb := ps.CreateMailbox()
	ps.SubOnce(mb, "t1")

	ps.Pub(&Message{Payload: "hi"}, "t1")
	c.Check(<-mb, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})

	_, ok := <-mb
	c.Check(ok, check.Equals, false)
	ps.Shutdown()
}

func (s *Suite) TestUnsub(c *check.C) {
	ps := New(1)
	mb := ps.CreateMailbox()
	ps.Sub(mb, "t1")

	ps.Pub(&Message{Payload: "hi"}, "t1")
	c.Check(<-mb, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})

	ps.Unsub(mb, "t1")
	_, ok := <-mb
	c.Check(ok, check.Equals, false)
	ps.Shutdown()
}

func (s *Suite) TestUnsubAll(c *check.C) {
	ps := New(1)
	mb1 := ps.CreateMailbox()
	mb2 := ps.CreateMailbox()

	ps.Sub(mb1, "t1", "t2", "t3")
	ps.Sub(mb2, "t1", "t3")

	ps.Unsub(mb1)

	m, ok := <-mb1
	c.Check(ok, check.Equals, false)

	ps.Pub(&Message{Payload: "hi"}, "t1")
	m, ok = <-mb2
	c.Check(m, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})

	ps.Shutdown()
}

func (s *Suite) TestClose(c *check.C) {
	ps := New(1)
	mb1 := ps.CreateMailbox()
	mb2 := ps.CreateMailbox()
	mb3 := ps.CreateMailbox()
	mb4 := ps.CreateMailbox()
	ps.Sub(mb1, "t1")
	ps.Sub(mb2, "t1")
	ps.Sub(mb3, "t2")
	ps.Sub(mb4, "t3")

	ps.Pub(&Message{Payload: "hi"}, "t1")
	ps.Pub(&Message{Payload: "hello"}, "t2")
	c.Check(<-mb1, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})
	c.Check(<-mb2, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})
	c.Check(<-mb3, check.DeepEquals, &Mailbox{Topic: "t2", Content: &Message{Payload: "hello"}})

	ps.Close("t1", "t2")
	_, ok := <-mb1
	c.Check(ok, check.Equals, false)
	_, ok = <-mb2
	c.Check(ok, check.Equals, false)
	_, ok = <-mb3
	c.Check(ok, check.Equals, false)

	ps.Pub(&Message{Payload: "welcome"}, "t3")
	c.Check(<-mb4, check.DeepEquals, &Mailbox{Topic: "t3", Content: &Message{Payload: "welcome"}})

	ps.Shutdown()
}

func (s *Suite) TestUnsubAfterClose(c *check.C) {
	ps := New(1)
	mb := ps.CreateMailbox()
	ps.Sub(mb, "t1")
	defer func() {
		ps.Unsub(mb, "t1")
		ps.Shutdown()
	}()

	ps.Close("t1")
	_, ok := <-mb
	c.Check(ok, check.Equals, false)
}

func (s *Suite) TestShutdown(c *check.C) {
	start := runtime.NumGoroutine()
	New(10).Shutdown()
	time.Sleep(1)
	c.Check(runtime.NumGoroutine(), check.Equals, start)
}

func (s *Suite) TestMultiSub(c *check.C) {
	ps := New(1)
	mb := ps.CreateMailbox()
	ps.Sub(mb, "t1", "t2")

	ps.Pub(&Message{Payload: "hi"}, "t1")
	c.Check(<-mb, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})

	ps.Pub(&Message{Payload: "hello"}, "t2")
	c.Check(<-mb, check.DeepEquals, &Mailbox{Topic: "t2", Content: &Message{Payload: "hello"}})

	ps.Shutdown()
	_, ok := <-mb
	c.Check(ok, check.Equals, false)
}

func (s *Suite) TestMultiSubOnce(c *check.C) {
	ps := New(1)
	mb := ps.CreateMailbox()
	ps.SubOnce(mb, "t1", "t2")

	ps.Pub(&Message{Payload: "hi"}, "t1")
	c.Check(<-mb, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})

	ps.Pub(&Message{Payload: "hello"}, "t2")

	_, ok := <-mb
	c.Check(ok, check.Equals, false)
	ps.Shutdown()
}

func (s *Suite) TestMultiPub(c *check.C) {
	ps := New(1)
	mb1 := ps.CreateMailbox()
	mb2 := ps.CreateMailbox()

	ps.Sub(mb1, "t1")
	ps.Sub(mb2, "t2")

	ps.Pub(&Message{Payload: "hi"}, "t2", "t1")
	c.Check(<-mb1, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})
	c.Check(<-mb2, check.DeepEquals, &Mailbox{Topic: "t2", Content: &Message{Payload: "hi"}})

	ps.Shutdown()
}

func (s *Suite) TestMultiUnsub(c *check.C) {
	ps := New(1)
	mb := ps.CreateMailbox()
	ps.Sub(mb, "t1", "t2", "t3")

	ps.Unsub(mb, "t1")

	ps.Pub(&Message{Payload: "hi"}, "t1")

	ps.Pub(&Message{Payload: "hello"}, "t2")
	c.Check(<-mb, check.DeepEquals, &Mailbox{Topic: "t2", Content: &Message{Payload: "hello"}})

	ps.Unsub(mb, "t2", "t3")
	_, ok := <-mb
	c.Check(ok, check.Equals, false)

	ps.Shutdown()
}

func (s *Suite) TestMultiClose(c *check.C) {
	ps := New(1)
	mb := ps.CreateMailbox()
	ps.Sub(mb, "t1", "t2")

	ps.Pub(&Message{Payload: "hi"}, "t1")
	c.Check(<-mb, check.DeepEquals, &Mailbox{Topic: "t1", Content: &Message{Payload: "hi"}})

	ps.Close("t1")
	ps.Pub(&Message{Payload: "hello"}, "t2")
	c.Check(<-mb, check.DeepEquals, &Mailbox{Topic: "t2", Content: &Message{Payload: "hello"}})

	ps.Close("t2")
	_, ok := <-mb
	c.Check(ok, check.Equals, false)

	ps.Shutdown()
}
