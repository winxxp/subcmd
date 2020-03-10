//Package subcmd
//
// +build windows

package subcmd

import (
	sysctx "context"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"testing"
	"time"
)

func TestSubCmd(t *testing.T) {
	Convey("SubCmd", t, func() {
		Convey("normal", func() {
			ctx, cancel := sysctx.WithCancel(sysctx.Background())
			defer cancel()

			cmd := New("ping",
				Args("localhost", "-n", "4"),
				QuitHandle(func() {
					t.Log("quit")
					cancel()
				}),
				LogHandle(func(s []byte) {
					t.Log(string(s))
				}),
			)

			err := cmd.Run(ctx)
			So(err, ShouldBeNil)
		})
		Convey("interrupt", func() {
			ctx, cancel := sysctx.WithTimeout(sysctx.Background(), time.Second*5)
			defer cancel()

			cmd := New("ping",
				Args("localhost", "-n", "100"),
				QuitHandle(func() {
					t.Log("quit 2")
					cancel()
				}),
				LogHandle(func(s []byte) {
					log.Print(string(s))
				}),
			)

			err := cmd.Run(ctx)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "exit status 1")
		})
	})
}
