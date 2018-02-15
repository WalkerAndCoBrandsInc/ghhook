package ghhook

import (
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-github/github"
)

type Response struct {
	StatusCode      int
	Headers         map[string]string
	Body            string
	IsBase64Encoded bool
}

type InputFn func(interface{}) (*Response, error)

var (
	Handlers = map[Event][]InputFn{}

	ErrorResponseFn   = DefaultErrorResponseFn
	SuccessResponseFn = DefaultSuccessResponseFn

	ErrNoGithubEventHeader = errors.New("ERROR: no X-GitHub-Event header")
)

func EventHandler(event Event, fn InputFn) {
	Handlers[event] = append(Handlers[event], fn)
}

func DefaultHandler(r *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	eventName, ok := r.Headers["X-GitHub-Event"]
	if !ok {
		return ErrorResponseFn(ErrNoGithubEventHeader)
	}

	fns, ok := Handlers[Event(eventName)]
	if !ok {
		log.Printf("Dropping unregistered event:%s", eventName)
		return SuccessResponseFn()
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

func DefaultSuccessResponseFn() (*events.APIGatewayProxyResponse, error) {
	return &events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 200,
	}, nil
}

func convertResponseToEventsResponse(r *Response) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode:      r.StatusCode,
		Headers:         r.Headers,
		Body:            r.Body,
		IsBase64Encoded: r.IsBase64Encoded,
	}
}
