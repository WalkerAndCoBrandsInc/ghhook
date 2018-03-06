package ghhook

import (
	"fmt"
	"testing"

	"github.com/google/go-github/github"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEventHandler(t *testing.T) {
	var fn InputFn = func(e interface{}) (*Response, error) {
		return nil, nil
	}

	Convey("It saves input fn to event name", t, func() {
		Reset(func() { ResetHandlers() })

		EventHandler(PullRequestEvent, fn)

		So(len(Handlers), ShouldEqual, 1)
		So(len(Handlers[PullRequestEvent]), ShouldEqual, 1)
		So(Handlers[PullRequestEvent][0], ShouldEqual, fn)
	})

	Convey("It appends input fn to event name", t, func() {
		Reset(func() { ResetHandlers() })

		EventHandler(PullRequestEvent, fn)
		EventHandler(PullRequestEvent, fn)

		So(len(Handlers[PullRequestEvent]), ShouldEqual, 2)
		So(Handlers[PullRequestEvent][0], ShouldEqual, fn)
		So(Handlers[PullRequestEvent][1], ShouldEqual, fn)
	})
}

func TestDefaultHandler(t *testing.T) {
	Convey("It drops unregistered events", t, func() {
		Reset(func() { ResetHandlers() })

		resp, err := DefaultHandler(PullRequestProxyRequest)
		So(err, ShouldBeNil)

		So(resp, ShouldNotBeNil)
		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Body, ShouldEqual, "")
	})

	Convey("It calls registered fn for event", t, func() {
		Reset(func() { ResetHandlers() })

		var fn InputFn = func(e interface{}) (*Response, error) {
			pr, ok := e.(*github.PullRequestEvent)
			So(ok, ShouldBeTrue)

			return &Response{
				Body:       fmt.Sprintf("%s", *pr.Action),
				StatusCode: 200,
			}, nil
		}

		EventHandler(PullRequestEvent, fn)

		resp, err := DefaultHandler(PullRequestProxyRequest)
		So(err, ShouldBeNil)

		So(resp, ShouldNotBeNil)
		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Body, ShouldEqual, "opened")
	})
}
