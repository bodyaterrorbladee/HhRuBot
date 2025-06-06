package hh

import(
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
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


func (c *Client) GetVacancies() ([]Vacancy,error){
	params := url.Values{}
	params.Set("text", "golang OR go developer OR golang developer")
	params.Add("area", "1")
	params.Add("area","2")
	params.Set("order_by", "publication_time")
	params.Set("per_page","20")
	params.Set("page","0")
	params.Set("only_with_salary","false")

	url := fmt.Sprintf("%s?%s",c.baseURL,params.Encode())


	req,err := http.NewRequest("GET",url,nil)
	if err!=nil{
		return nil,err
	}

	req.Header.Set("User-agent","go-job-bot/0.1")

	resp, err := c.client.Do(req)
	if err!=nil{
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hh.ru error: %s", string(body))
	}
	var data ResponseHH
	if err := json.NewDecoder(resp.Body).Decode(&data); err!=nil{
		return nil, err
	}
	return data.Items, nil
}