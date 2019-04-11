// Copyright 2018 The go-zeromq Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zmq4_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	//"github.com/go-zeromq/zmq4"
	"github.com/itsblocked/zmq4"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	pushpulls = []testCasePushPull{
		{
			name:     "tcp-push-pull",
			endpoint: must(EndPoint("tcp")),
			push:     zmq4.NewPush(bkg),
			pull:     zmq4.NewPull(bkg),
		},
		{
			name:     "ipc-push-pull",
			endpoint: "ipc://ipc-push-pull",
			push:     zmq4.NewPush(bkg),
			pull:     zmq4.NewPull(bkg),
		},
		{
			name:     "inproc-push-pull",
			endpoint: "inproc://push-pull",
			push:     zmq4.NewPush(bkg),
			pull:     zmq4.NewPull(bkg),
		},
	}
)

type testCasePushPull struct {
	name     string
	skip     bool
	endpoint string
	push     zmq4.Socket
	pull     zmq4.Socket
}

func TestPushPull(t *testing.T) {
	var (
		hello = zmq4.NewMsg([]byte("HELLO WORLD"))
		bye   = zmq4.NewMsgFrom([]byte("GOOD"), []byte("BYE"))
	)

	for i := range pushpulls {
		tc := pushpulls[i]
		t.Run(tc.name, func(t *testing.T) {
			defer tc.pull.Close()
			defer tc.push.Close()

			if tc.skip {
				t.Skipf(tc.name)
			}
			t.Parallel()

			ep := tc.endpoint
			cleanUp(ep)

			ctx, timeout := context.WithTimeout(context.Background(), 20*time.Second)
			defer timeout()

			grp, ctx := errgroup.WithContext(ctx)
			grp.Go(func() error {

				err := tc.push.Listen(ep)
				if err != nil {
					return errors.Wrapf(err, "could not listen")
				}

				err = tc.push.Send(hello)
				if err != nil {
					return errors.Wrapf(err, "could not send %v", hello)
				}

				err = tc.push.Send(bye)
				if err != nil {
					return errors.Wrapf(err, "could not send %v", bye)
				}
				return err
			})
			grp.Go(func() error {

				err := tc.pull.Dial(ep)
				if err != nil {
					return errors.Wrapf(err, "could not dial")
				}

				msg, err := tc.pull.Recv()
				if err != nil {
					return errors.Wrapf(err, "could not recv %v", hello)
				}

				if got, want := msg, hello; !reflect.DeepEqual(got, want) {
					return errors.Errorf("recv1: got = %v, want= %v", got, want)
				}

				msg, err = tc.pull.Recv()
				if err != nil {
					return errors.Wrapf(err, "could not recv %v", bye)
				}

				if got, want := msg, bye; !reflect.DeepEqual(got, want) {
					return errors.Errorf("recv2: got = %v, want= %v", got, want)
				}

				return err
			})
			if err := grp.Wait(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
