package metrics

import "context"

type Recorder interface {
	CountHTTPRequest(ctx context.Context, method, path string, status int)
}

type NoopRecorder struct{}

func (NoopRecorder) CountHTTPRequest(context.Context, string, string, int) {}
