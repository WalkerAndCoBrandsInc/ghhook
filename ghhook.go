package ghhook

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-github/github"
)

// Response is returned by InputFn. It's exactly same as
// events.APIGatewayProxyResponse, but is required so there's no conflict when
// the caller of this library also uses the events library.
type Response struct {
	StatusCode      int
	Headers         map[string]string
	Body            string
	IsBase64Encoded bool
}

// InputFn is the type passed to event handler.
type InputFn func(interface{}) (*Response, error)

var (
	// Handlers is global list of the webhook functions mapped to their respective
	// webhook event names.
	Handlers = map[Event][]InputFn{}

	// ErrorResponseFn is used by DefaultHandler to return error responses.
	ErrorResponseFn = DefaultErrorResponseFn

	// SuccessResponseFn is used by DefaultHandler to return success responses for
	// webhooks that don't have a InputFn assocation with them.
	SuccessResponseFn = DefaultSuccessResponseFn

	// ErrNoGithubEventHeader is return when request header does not contain the
	// required header.
	//
	// Event names are sent in the header under 'X-GitHub-Event' by Github.
	ErrNoGithubEventHeader = errors.New("ERROR: no 'X-GitHub-Event' header")
)

// EventHandler appends the given InputFn to the given event.
//
// Example:
//	ghhook.EventHandler(ghhook.PullRequestEvent, func(e interface{}) (*ghhook.Response, error) {
//		pr, _ := e.(*github.PullRequestEvent)
//
//		return &ghhook.Response{
//			Body:       fmt.Sprintf("%s", *pr.Action),
//			StatusCode: 200,
//		}, nil
//	})
func EventHandler(event Event, fn InputFn) {
	Handlers[event] = append(Handlers[event], fn)
}

// EventHandlerActionFilter is similar to EventHandler with addition of checking
// if the top level keys match the given filters.
//
// Example:
//  # only listens to 'opened' action.
//	ghhook.EventHandlerActionFilter(
//	  ghhook.PullRequestEvent,
//	  map[string][]string{"action": []string{"opened"}},
//	  func(e interface{}) (*ghhook.Response, error) { return nil, nil },
//	)
func EventHandlerActionFilter(event Event, filters map[string][]string, fn InputFn) {
	wrappedFn := func(i interface{}) (*Response, error) {
		b, err := json.Marshal(i)
		if err != nil {
			return nil, err
		}

		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}

		for key, allowedValues := range filters {
			value, ok := m[key]
			if !ok {
				return localSuccessResp(fmt.Sprintf("No key:'%s' in event body", key))
			}

			for _, allowed := range allowedValues {
				if value == allowed {
					return fn(i)
				}
			}
		}

		return localSuccessResp(fmt.Sprintf("Dropping unregistered value: '%v'", filters))
	}

	Handlers[event] = append(Handlers[event], wrappedFn)
}

// EventHandlerFunctionFilter is similar to EventHandler with addition of
// checking if event matches the given filter function.
//
// Example:
//  # only listens to 'opened' action.
//	ghhook.EventHandlerFunctionFilter(
//	  ghhook.PullRequestEvent,
//    func(e map[string]interface{}) bool { return true },
//	  func(e interface{}) (*ghhook.Response, error) { return nil, nil },
//	)
func EventHandlerFunctionFilter(event Event, filterFn func(map[string]interface{}) bool, fn InputFn) {
	wrappedFn := func(i interface{}) (*Response, error) {
		b, err := json.Marshal(i)
		if err != nil {
			return nil, err
		}

		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}

		if !filterFn(m) {
			return localSuccessResp("Dropping unmatched event for function")
		}

		return fn(i)
	}

	Handlers[event] = append(Handlers[event], wrappedFn)
}

// DefaultHandler is a Lambda compatible handler that receives
// APIGatewayProxyRequest, ie Github webhook and calls the InputFn mapped to the
// event name.
//
// If there are multiple InputFn for event, if all are successful only the last
// response is returned, but if any of them fails, it stops execution and
// returns the error.
func DefaultHandler(r *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	eventName, ok := r.Headers["X-GitHub-Event"]
	if !ok {
		return ErrorResponseFn(ErrNoGithubEventHeader)
	}

	fns, ok := Handlers[Event(eventName)]
	if !ok {
		return SuccessResponseFn(fmt.Sprintf("Dropping unregistered event: '%s'", eventName))
	}

	i, err := github.ParseWebHook(eventName, []byte(r.Body))
	if err != nil {
		return ErrorResponseFn(err)
	}

	var lastResponse *Response
	for _, fn := range fns {
		var err error
		if lastResponse, err = fn(i); err != nil {
			return ErrorResponseFn(err)
		}
	}

	return convertResponseToEventsResponse(lastResponse), nil
}

func DefaultErrorResponseFn(err error) (*events.APIGatewayProxyResponse, error) {
	return &events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: 500,
	}, nil
}

func DefaultSuccessResponseFn(body string) (*events.APIGatewayProxyResponse, error) {
	return &events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 200,
	}, nil
}

func localSuccessResp(body string) (*Response, error) {
	return &Response{
		Body:       body,
		StatusCode: 200,
	}, nil
}

// ResetHandlers is used to clear out the handlers. This is mainly to be used in tests.
func ResetHandlers() {
	Handlers = map[Event][]InputFn{}
}

func convertResponseToEventsResponse(r *Response) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode:      r.StatusCode,
		Headers:         r.Headers,
		Body:            r.Body,
		IsBase64Encoded: r.IsBase64Encoded,
	}
}
