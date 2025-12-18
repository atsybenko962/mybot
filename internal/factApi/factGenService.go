package factApi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Fact struct {
	Text string `json:"text"`
}

type FactAPI struct {
	client *http.Client
}

func NewFactAPI() *FactAPI {
	return &FactAPI{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetRandomFact получает случайный факт и переводит его на русский язык.
func (f *FactAPI) GetRandomFact(ctx context.Context) (*Fact, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://uselessfacts.jsph.pl/api/v2/facts/random", nil)
	if err != nil {
		return nil, errors.New("ошибка создания запроса")
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, errors.New("ошибка запроса к API")
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("API вернул неожиданный статус")
	}

	var factResponse Fact

	if err := json.NewDecoder(resp.Body).Decode(&factResponse); err != nil {
		return nil, errors.New("ошибка декодирования JSON")
	}

	if factResponse.Text == "" {
		return nil, errors.New("получен пустой факт")
	}

	t, err := translateENtoRU(factResponse.Text)
	if err != nil {
		return nil, errors.New("ошибка перевода")
	}

	factResponse.Text = t

	return &factResponse, nil
}
