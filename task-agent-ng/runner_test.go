package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQueueRun(t *testing.T) {
	result, err := QueueRun("sample")
	if result != 0 {
		t.Errorf("Task failed to queue: %s", err)
	}
}

func TestRunTask(t *testing.T) {
	Convey("When a valid command is executed", t, func() {
		exitStatus, _ := RunTask("ls", "/", 30)
		Convey("The exitStatus should be zero", func() {
			So(exitStatus, ShouldEqual, 0)
		})
	})
	Convey("When an invalid command is executed", t, func() {
		exitStatus, _ := RunTask("invalid_ls", "/", 30)
		Convey("The exitStatus should not equal zero", func() {
			So(exitStatus, ShouldNotEqual, 0)
		})
	})
	Convey("When a command runs too long", t, func() {
		exitStatus, _ := RunTask("sleep", "30", 3)
		Convey("The process should be killed", func() {
			So(exitStatus, ShouldEqual, 3)
		})
	})
}
