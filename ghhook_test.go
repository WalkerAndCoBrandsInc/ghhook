package ghhook

import (
	"fmt"
	"testing"

	"github.com/google/go-github/github"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEventHandler(t *testing.T) {
	Convey("EventHandler", t, func() {
		var fn InputFn = func(e interface{}) (*Response, error) {
			return nil, nil
		}

		Reset(func() { ResetHandlers() })

		Convey("It saves input fn to event name", func() {
			EventHandler(PullRequestEvent, fn)

			So(len(Handlers), ShouldEqual, 1)
			So(len(Handlers[PullRequestEvent]), ShouldEqual, 1)
			So(Handlers[PullRequestEvent][0], ShouldEqual, fn)
		})

		Convey("It appends input fn to event name", func() {
			EventHandler(PullRequestEvent, fn)
			EventHandler(PullRequestEvent, fn)

			So(len(Handlers[PullRequestEvent]), ShouldEqual, 2)
			So(Handlers[PullRequestEvent][0], ShouldEqual, fn)
			So(Handlers[PullRequestEvent][1], ShouldEqual, fn)
		})
	})
}

func TestDefaultHandler(t *testing.T) {
	Convey("DefaultHandler", t, func() {
		Reset(func() { ResetHandlers() })

		var fn InputFn = func(e interface{}) (*Response, error) {
			pr, ok := e.(*github.PullRequestEvent)
			So(ok, ShouldBeTrue)

			return &Response{
				Body:       fmt.Sprintf("%s", *pr.Action),
				StatusCode: 200,
			}, nil
		}

		Convey("It drops unregistered events", func() {
			resp, err := DefaultHandler(PullRequestProxyRequest)
			So(err, ShouldBeNil)

			So(resp, ShouldNotBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Body, ShouldEqual, "Dropping unregistered event: 'pull_request'")
		})

		Convey("It drops unregistered actions", func() {
			EventHandlerActionFilter(PullRequestEvent, []string{"reopened"}, fn)

			resp, err := DefaultHandler(PullRequestProxyRequest)
			So(err, ShouldBeNil)
			So(resp.Body, ShouldEqual, "Dropping unregistered action: 'opened'")
		})

		Convey("It calls registered fn for event", func() {
			EventHandler(PullRequestEvent, fn)

			resp, err := DefaultHandler(PullRequestProxyRequest)
			So(err, ShouldBeNil)

			So(resp, ShouldNotBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Body, ShouldEqual, "opened")
		})

		Convey("It calls registered fn for event with action filter", func() {
			EventHandlerActionFilter(PullRequestEvent, []string{"opened"}, fn)

			resp, err := DefaultHandler(PullRequestProxyRequest)
			So(err, ShouldBeNil)

			So(resp, ShouldNotBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Body, ShouldEqual, "opened")
		})
	})
}
