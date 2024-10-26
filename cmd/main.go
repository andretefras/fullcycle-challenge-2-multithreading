package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type APIResponse struct {
	URL         string
	Response    map[string]interface{}
	ElapsedTime time.Duration
	Success     bool
}

func fetchAPI(url string, wg *sync.WaitGroup, ch chan<- APIResponse) {
	defer wg.Done()
	start := time.Now()

	client := http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		ch <- APIResponse{URL: url, Success: false}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- APIResponse{URL: url, Success: false}
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		ch <- APIResponse{URL: url, Success: false}
		return
	}

	elapsed := time.Since(start)
	ch <- APIResponse{URL: url, Response: result, ElapsedTime: elapsed, Success: true}
}

func getApisUrls(cep string) []string {
	return []string{
		fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep),
		fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep),
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Digite o CEP: ")
	cep, _ := reader.ReadString('\n')
	cep = cep[:len(cep)-1]

	apis := getApisUrls(cep)

	var wg sync.WaitGroup
	ch := make(chan APIResponse, len(apis))

	for _, api := range apis {
		wg.Add(1)
		go fetchAPI(api, &wg, ch)
	}

	wg.Wait()
	close(ch)

	var fastest APIResponse
	for response := range ch {
		if response.Success {
			if !fastest.Success || response.ElapsedTime < fastest.ElapsedTime {
				fastest = response
			}
		}
	}

	if fastest.Success {
		fmt.Printf("API mais rápida: %s\nEndereço: %v\n", fastest.URL, fastest.Response)
	} else {
		fmt.Println("Nenhuma das APIs respondeu dentro do tempo limite de 1 segundo.")
	}
}
