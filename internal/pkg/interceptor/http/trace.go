// gen by iyfiysi at 2021 May 19

package http

import (
	"context"
	"iyfiysi.com/short_url/internal/pkg/trace"
	"net/http"
)

// Trace trace for http
func Trace(
	next func(w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		url := r.URL.String()
		httpSpan, ctx := trace.StartTrace(ctx, url)
		defer httpSpan.Finish()
		r = r.WithContext(ctx)
		next(w, r)
	}
}
