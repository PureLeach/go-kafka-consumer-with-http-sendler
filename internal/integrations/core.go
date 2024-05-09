package integrations

import (
	"encoding/json"
	"events_consumer/internal/config"
	"events_consumer/internal/models"
	"events_consumer/internal/utils"
	"io"
	"log"
	"strconv"
	"sync"

	"github.com/jellydator/ttlcache/v3"

	"net/http"
	"net/url"
)

func LoadCoreVehicle(cfg *config.Config) {

	urlVehicleCore, err := url.JoinPath(cfg.CORE_HOST, cfg.CORE_VEHICLE_SERVICE_PATH)
	if err != nil {
		log.Fatal(err)
	}

	client := utils.CreateClient()

	pages, err := checkCoreApi(urlVehicleCore, cfg.CORE_API_KEY, client)
	if err != nil {
		log.Fatal(err)
	}

	// syncSaveVehiclesToCache(urlVehicleCore, cfg.CORE_API_KEY, int(pages), client)
	asyncSaveVehiclesToCache(urlVehicleCore, cfg.CORE_API_KEY, int(pages), client)
}

func checkCoreApi(url string, apiKey string, client *http.Client) (int64, error) {
	log.Printf("Trying to get all vehicles from the core service. url = %s\n", url)

	coreResponse, err := getCoreVehicles(url, apiKey, nil, client)
	if err != nil {
		log.Fatal(err)
	}

	pages := coreResponse.Result.Pages
	return pages, nil
}

func asyncSaveVehiclesToCache(url string, apiKey string, pages int, client *http.Client) {
	var wg sync.WaitGroup
	wg.Add(pages)

	results := make(chan models.CoreResponse, pages)
	errors := make(chan error, pages)

	for i := 1; i <= pages; i++ {
		go func(page int) {
			defer wg.Done()

			coreResponse, err := getCoreVehicles(url, apiKey, &page, client)
			if err != nil {
				errors <- err
				return
			}
			results <- coreResponse
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	for coreResponse := range results {
		for _, item := range coreResponse.Result.Items {
			utils.CacheMain.Set(item.KmclVehicleID, item.ID, ttlcache.NoTTL)
			log.Printf("Saved the vehicle in the cache. kmcl_vehicle_id = %s, core_vehicle_id = %s\n", item.KmclVehicleID, item.ID)
		}
	}

	for err := range errors {
		log.Fatalln("Error saving the vehicle in the cache:", err)
	}

}

func syncSaveVehiclesToCache(url string, apiKey string, pages int, client *http.Client) {
	for i := 1; i <= pages; i++ {
		coreResponse, err := getCoreVehicles(url, apiKey, &i, client)
		if err != nil {
			log.Fatalln("Error saving the vehicle in the cache:", err)
		}
		for _, item := range coreResponse.Result.Items {
			utils.CacheMain.Set(item.KmclVehicleID, item.ID, ttlcache.NoTTL)
			log.Printf("Saved the vehicle in the cache. kmcl_vehicle_id = %s, core_vehicle_id = %s\n", item.KmclVehicleID, item.ID)
		}
	}

}

func getCoreVehicles(url string, apiKey string, page *int, client *http.Client) (models.CoreResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating GET request:", err)
	}

	req.Header.Set("X-Api-Key", apiKey)

	if page != nil {
		q := req.URL.Query()
		q.Add("page", strconv.Itoa(*page))
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending GET request:", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	log.Println("Response Status Code:", statusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading the response: %v", err)
	}

	var coreResponse models.CoreResponse
	coreErr := json.Unmarshal(body, &coreResponse)
	if coreErr != nil {
		log.Printf("Error parsing the message: msg = %s, err = %v", string(body), coreErr)
	}
	return coreResponse, nil
}
