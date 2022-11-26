package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"gophermart/internal/app/domain"
	"io"
	"net/http"
	"strconv"
	"time"
)

const retryAfterTime = time.Second * 60

type AccrualCalculationService struct {
	accrualSystemAddr string
}

func NewAccrualCalculationService(accrualSystemAddr string) *AccrualCalculationService {
	return &AccrualCalculationService{accrualSystemAddr: accrualSystemAddr}
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
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respStatusCode := resp.StatusCode
	// 409 статус может быть в случае, если заказ ранее уже был создан в системе начисления
	if respStatusCode != http.StatusAccepted && respStatusCode != http.StatusConflict {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = resp.Body.Close()
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)
		return fmt.Errorf("creating order for accrual calculation failed: status code - %d, body - %v", respStatusCode, bodyString)
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
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error().Msg(fmt.Sprintf("error on response body close: %v", err.Error()))
		}
	}()

	respStatusCode := resp.StatusCode
	if respStatusCode == http.StatusTooManyRequests {
		<-time.After(retryAfterTime)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.Error().Msg(fmt.Sprintf("error on response body close: %v", err.Error()))
			}
		}()
	}
	if respStatusCode != http.StatusOK {
		errMsg := "request to accrual system failed: status of response - " + strconv.Itoa(respStatusCode)
		log.Error().Msg(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	var accrualRes domain.AccrualCalculationRes
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&accrualRes)
	if err != nil {
		log.Error().Msg("request to accrual system failed: " + err.Error())
		return nil, err
	}
	log.Info().Msg(fmt.Sprintf("accrual result is: %v", accrualRes))

	return &accrualRes, nil
}
