package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
	"net/http"
	"strconv"
)

type RequestsWorker interface {
	HandleRequest(ctx context.Context, req *http.Request) (*domain.ResponseWithReadBody, error)
}

type AccrualCalculationService struct {
	accrualSystemAddr string
	requestsWorker    RequestsWorker
}

func NewAccrualCalculationService(accrualSystemAddr string, worker RequestsWorker) *AccrualCalculationService {
	return &AccrualCalculationService{accrualSystemAddr: accrualSystemAddr, requestsWorker: worker}
}

type orderInput struct {
	Order string `json:"order"`
}

func (s *AccrualCalculationService) CreateOrderForCalculation(orderNumber string) error {
	requestURL := s.accrualSystemAddr + "/api/orders"
	reqBody, err := json.Marshal(&orderInput{Order: orderNumber})
	reqBodyReader := bytes.NewReader(reqBody)
	if err != nil {
		return err
	}

	log.Info().Msg(fmt.Sprintf("making request to %s", requestURL))
	req, err := http.NewRequest(http.MethodPost, requestURL, reqBodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := s.requestsWorker.HandleRequest(context.Background(), req)
	if err != nil {
		return err
	}

	respStatusCode := res.Response.StatusCode
	if respStatusCode != http.StatusAccepted {
		respBody := string(res.ReadBody)
		return fmt.Errorf("creating order for accrual calculation failed: status code - %d, body - %v", respStatusCode, respBody)
	}

	return nil
}

func (s *AccrualCalculationService) GetOrderAccrualRes(orderNumber string) (*domain.AccrualCalculationRes, error) {
	requestURL := s.accrualSystemAddr + "/api/orders/"
	// проверяем, был ли заказ обработан
	req, err := http.NewRequest(http.MethodGet, requestURL+orderNumber, nil)
	if err != nil {
		log.Error().Msg("request to accrual system failed: " + err.Error())
		return nil, err
	}
	res, err := s.requestsWorker.HandleRequest(context.Background(), req)
	if err != nil {
		log.Error().Msg("request to accrual system failed: " + err.Error())
		return nil, err
	}
	respStatusCode := res.Response.StatusCode
	if respStatusCode != http.StatusOK {
		errMsg := "request to accrual system failed: status of response - " + strconv.Itoa(respStatusCode)
		log.Error().Msg(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	var accrualRes domain.AccrualCalculationRes
	log.Info().Msg(fmt.Sprintf("accrual result is: %v", res.ReadBody))
	err = json.NewDecoder(bytes.NewReader(res.ReadBody)).Decode(&accrualRes)
	if err != nil {
		log.Error().Msg("request to accrual system failed: " + err.Error())
		return nil, err
	}

	return &accrualRes, nil
}
