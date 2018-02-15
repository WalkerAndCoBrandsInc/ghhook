# ghhook

ghhook is a Go library that makes working with Github webhooks delivered with AWS APIGateway easier.

At its core you register an event name with a function that's run whenever webhook with that event name is received.

## Example:

```Go
// Listen to "pull_request" event and return with 200 status.
ghhook.EventHandler(ghhook.PullRequestEvent, func(e interface{}) (*ghhook.Response, error) {
  pr, _ := e.(*github.PullRequestEvent)

  return &ghhook.Response{
    Body:       fmt.Sprintf("%s", *pr.Action),
    StatusCode: 200,
  }, nil
})
```

The input function has to be interface `func(e interface{}) (*ghhook.Response, error)`. `e` maps to event struct that's specific to the webhook. ghhook uses the webhook event definitions from [google/go-github](https://github.com/google/go-github). See [here](https://github.com/google/go-github/blob/df47db1628185875602e66d3356ae7337b52bba3/github/activity_events.go#L35) for list of all the available events and their respective mapping.

## DefaultHandler

DefaultHandler is the start point which receives the webhook and runs the registered functions for the webhook. It's possible to use it as a lambda function out of the box.

```Go
lambda.Start(ghhook.DefaultHandler)
```
