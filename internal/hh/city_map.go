package hh

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var cityMap = make(map[string]string)

type Area struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Areas  []Area  `json:"areas"`
}

// InitCityMap загружает список всех городов и строит карту для поиска
func InitCityMap() error {
	resp, err := http.Get("https://api.hh.ru/areas")
	if err != nil {
		return fmt.Errorf("ошибка запроса городов HH: %w", err)
	}
	defer resp.Body.Close()

	var countries []Area
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		return fmt.Errorf("ошибка декодирования JSON: %w", err)
	}

	for _, country := range countries {
		walkAreas(country)
	}

	return nil
}

// walkAreas рекурсивно обходит структуру регионов и наполняет cityMap
func walkAreas(area Area) {
	lowerName := strings.ToLower(area.Name)
	if _, exists := cityMap[lowerName]; !exists && area.ID != "" {
		cityMap[lowerName] = area.ID
	}

	for _, subArea := range area.Areas {
		walkAreas(subArea)
	}
}

// CityToAreaID возвращает код региона для города
func CityToAreaID(name string) string {
	return cityMap[strings.ToLower(strings.TrimSpace(name))]
}
