package factApi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type MyMemoryResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
	QuotaFinished  bool     `json:"quotaFinished"`
	MTLang         []string `json:"mtLang"` // ["en", "ru"]
	ResponseStatus int      `json:"responseStatus"`
}

// translateENtoRU переводит текст с английского на русский через MyMemory API.
// email — опционально, но рекомендуется для надёжности.
func translateENtoRU(text string) (string, error) {
	baseURL := "https://api.mymemory.translated.net/get"
	params := url.Values{}
	params.Add("q", text)
	params.Add("langpair", "en|ru")

	reqURL := baseURL + "?" + params.Encode()

	resp, err := http.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("ошибка HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result MyMemoryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("ошибка парсинга JSON: %w\nRaw: %s", err, string(body))
	}

	if result.ResponseStatus != 200 {
		return "", fmt.Errorf("API error %d", result.ResponseStatus)
	}

	if result.QuotaFinished {
		return "", fmt.Errorf("квота исчерпана (1000 запросов/день)")
	}

	return result.ResponseData.TranslatedText, nil
}
