package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
	"io"
	"net/http"
	"strconv"
)

type RequestsWorker interface {
	HandleRequest(ctx context.Context, req *http.Request) (*http.Response, error)
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
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}
	res, err := s.requestsWorker.HandleRequest(context.Background(), req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusAccepted {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		respBody := string(bodyBytes)
		return fmt.Errorf("creating order for accrual calculation failed: status code - %d, body - %v", res.StatusCode, respBody)
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
	if res.StatusCode != http.StatusOK {
		log.Error().Msg("request to accrual system failed: status of response - " + strconv.Itoa(res.StatusCode))
		return nil, err
	}

	var accrualRes domain.AccrualCalculationRes
	log.Info().Msg(fmt.Sprintf("accrual result is: %v", res.Body))
	err = json.NewDecoder(res.Body).Decode(&accrualRes)
	if err != nil {
		log.Error().Msg("request to accrual system failed: " + err.Error())
		return nil, err
	}

	return &accrualRes, nil
}
