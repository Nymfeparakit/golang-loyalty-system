package workers

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
	"io"
	"net/http"
	"sync"
	"time"
)

const reqsRate = 1 * time.Second
const respChTimeout = 5 * time.Second

type RequestWithResponseCh struct {
	ctx    context.Context
	req    *http.Request
	respCh chan *domain.ResponseWithReadBody
}

// RateLimitedReqWorker получает и затем отправляет переданные ему запросы с учетом заданного лимита на эти запросы
type RateLimitedReqWorker struct {
	reqCh chan RequestWithResponseCh
}

func NewRateLimitedReqWorker() *RateLimitedReqWorker {
	reqCh := make(chan RequestWithResponseCh)
	return &RateLimitedReqWorker{reqCh: reqCh}
}

func (w *RateLimitedReqWorker) HandleRequest(ctx context.Context, req *http.Request) (*domain.ResponseWithReadBody, error) {
	log.Info().Msg(fmt.Sprintf("starting handling new request: %v", req))
	// создаем канал, в котором будет ожидать ответа на текущий request
	respCh := make(chan *domain.ResponseWithReadBody)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	w.reqCh <- RequestWithResponseCh{req: req, respCh: respCh, ctx: ctx}

	var resp *domain.ResponseWithReadBody
	select {
	case r := <-respCh:
		resp = r
	case <-time.After(respChTimeout):
		return nil, fmt.Errorf("response channel timout exceeded")
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return resp, nil
}

func (w *RateLimitedReqWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Info().Msg("reqs worker: waiting for new requests")
	ticker := time.NewTicker(reqsRate)
	defer ticker.Stop()
	for {
		select {
		case req := <-w.reqCh:
			<-ticker.C
			go w.executeRequest(req)
		case <-ctx.Done():
			return
		}
	}
}

// todo: тут нужно что-то менять с контекстом?
func (w *RateLimitedReqWorker) executeRequest(req RequestWithResponseCh) {
	defer close(req.respCh)
	client := http.DefaultClient
	log.Info().Msg(fmt.Sprintf("executing request: %v", req))
	resp, err := client.Do(req.req)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to make request to AS: %v", err.Error()))
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to make request to AS: %v", err.Error()))
		return
	}
	responseWithBody := &domain.ResponseWithReadBody{ReadBody: bodyBytes, Response: resp}

	select {
	// если контекст был отменен, то отправлять ответ уже не нужно
	case <-req.ctx.Done():
		return
	default:
		req.respCh <- responseWithBody
	}
}
