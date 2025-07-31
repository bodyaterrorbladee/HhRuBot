package hh

import(
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"strings"
)


type Vacancy struct{
	Id string
	Name string
	Url string
	Area struct{
		Name string	`json:"name"`
	}`json:"area"`
}

type ResponseHH struct{
	Items []Vacancy `json:"items"`
}

type Client struct{
	baseURL string
	client *http.Client
}

func NewClient() *Client{
	return &Client{
		baseURL: "https://api.hh.ru/vacancies",
		client: &http.Client{
			Timeout: 10 *time.Second,
		} ,
	}
}


func (c *Client) GetVacancies(tags []string, cities []string) ([]Vacancy, error)  {
	params := url.Values{}

	// Формируем строку поиска из тегов
	if len(tags) > 0 {
		searchText := strings.Join(tags, " OR ")
		params.Set("text", searchText)
	} else {
		params.Set("text", "golang") // fallback
	}

	// Добавляем регионы (area) — по коду HH
	if len(cities) > 0 {
		for _, city := range cities {
			areaCode := mapCityToAreaCode(strings.ToLower(strings.TrimSpace(city)))
			if areaCode != "" {
				params.Add("area", areaCode)
			}
		}
	} else {
		// fallback: Москва и Питер
		params.Add("area", "1")
		params.Add("area", "2")
	}

	params.Set("order_by", "publication_time")
	params.Set("per_page", "20")
	params.Set("page", "0")
	params.Set("only_with_salary", "false")

	url := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-agent", "go-job-bot/0.1")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hh.ru error: %s", string(body))
	}

	var data ResponseHH
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Items, nil
}
func mapCityToAreaCode(city string) string {
	switch city {
	case "москва":
		return "1"
	case "санкт-петербург", "питер", "спб":
		return "2"
	case "екатеринбург":
		return "3"
	case "новосибирск":
		return "4"
	case "казань":
		return "88"
	default:
		return ""
	}
}