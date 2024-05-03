package integrations

import (
	"crypto/tls"
	"encoding/json"
	"events_consumer/internal/config"
	"events_consumer/internal/models"
	"events_consumer/internal/utils"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"

	"github.com/jellydator/ttlcache/v3"

	"net/http"
	"net/url"
)

func LoadCoreVehicle(cfg *config.Config) {

	urlVehicleCore, err := url.JoinPath(cfg.CORE_HOST, cfg.CORE_VEHICLE_SERVICE_PATH)
	if err != nil {
		log.Fatal(err)
	}

	// Создайте новый Transport с отключенной проверкой SSL
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Создайте клиент HTTP с настроенным Transport
	client := &http.Client{Transport: tr}

	pages, err := checkCoreApi(urlVehicleCore, cfg.CORE_API_KEY, client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("pages: %#v Type: %v\n", pages, reflect.TypeOf(pages))

	saveVehiclesToCache(urlVehicleCore, cfg.CORE_API_KEY, int(pages), client)
}

func checkCoreApi(url string, apiKey string, client *http.Client) (int64, error) {
	println("Попытка получить все машины из сервиса core url = %s", url)

	coreResponse, err := getCoreVehicles(url, apiKey, nil, client)
	if err != nil {
		log.Fatal(err)
	}

	pages := coreResponse.Result.Pages
	return pages, nil
}

func saveVehiclesToCache(url string, apiKey string, pages int, client *http.Client) {
	// for i := 1; i <= 1; i++ {
	for i := 1; i <= pages; i++ {
		coreResponse, err := getCoreVehicles(url, apiKey, &i, client)
		if err != nil {
			log.Fatal(err)
		}
		for _, item := range coreResponse.Result.Items {
			utils.CacheMain.Set(item.KmclVehicleID, item.ID, ttlcache.NoTTL)
			fmt.Printf("Сохранили машину в кэш kmcl_vehicle_id = %s, core_vehicle_id = %s\n", item.KmclVehicleID, item.ID)
		}
	}

}

func getCoreVehicles(url string, apiKey string, page *int, client *http.Client) (models.CoreResponse, error) {

	// Создайте новый запрос GET с параметрами
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating GET request:", err)
	}

	// Добавьте заголовок "X-Api-Key"
	req.Header.Set("X-Api-Key", apiKey)

	if page != nil {
		// Добавьте параметр "page"
		q := req.URL.Query()
		q.Add("page", strconv.Itoa(*page))
		req.URL.RawQuery = q.Encode()
	}

	// Выполните запрос
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending GET request:", err)
	}
	defer resp.Body.Close()

	// Прочитайте тело ответа, если это необходимо
	// Например, для прочтения JSON ответа можно использовать json.Decoder

	// Получите код статуса ответа
	statusCode := resp.StatusCode
	fmt.Println("Response Status Code:", statusCode)

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка при чтении ответа: %v", err)
	}

	var coreResponse models.CoreResponse
	coreErr := json.Unmarshal(body, &coreResponse)
	if coreErr != nil {
		log.Printf("Ошибка при парсинге сообщения: msg = %s, err = %v", string(body), coreErr)
	}
	return coreResponse, nil
}
