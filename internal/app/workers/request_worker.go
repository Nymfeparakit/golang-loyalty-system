package workers

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

const reqsRate = 1 * time.Second
const respChTimeout = 5 * time.Second

type RequestWithResponseCh struct {
	ctx    context.Context
	req    *http.Request
	respCh chan *http.Response
}

type RateLimitedReqWorker struct {
	reqCh chan RequestWithResponseCh
}

func NewRateLimitedReqWorker() *RateLimitedReqWorker {
	// todo: should be buffered?
	reqCh := make(chan RequestWithResponseCh)
	return &RateLimitedReqWorker{reqCh: reqCh}
}

func (w *RateLimitedReqWorker) HandleRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	log.Info().Msg(fmt.Sprintf("starting handling new request: %v", req))
	respCh := make(chan *http.Response)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	w.reqCh <- RequestWithResponseCh{req: req, respCh: respCh, ctx: ctx}
	defer close(respCh)

	var resp *http.Response
	select {
	case r := <-respCh:
		resp = r
	case <-time.After(respChTimeout):
		return nil, fmt.Errorf("response channel timout exceeded")
	}

	return resp, nil
}

func (w *RateLimitedReqWorker) ProcessRequests() {
	log.Info().Msg("reqs worker: waiting for new requests")
	throttle := time.Tick(reqsRate)
	for {
		req := <-w.reqCh
		<-throttle
		go w.executeRequest(req)
	}
}

func (w *RateLimitedReqWorker) executeRequest(req RequestWithResponseCh) {
	client := http.DefaultClient
	log.Info().Msg(fmt.Sprintf(fmt.Sprintf("executing request: %v", req)))
	resp, err := client.Do(req.req)
	if err != nil {
		// todo: what to do here?
		log.Error().Msg(fmt.Sprintf("failed to make request to AS: %v", err.Error()))
	}

	select {
	case <-req.ctx.Done():
		return
	default:
		req.respCh <- resp
	}
}
