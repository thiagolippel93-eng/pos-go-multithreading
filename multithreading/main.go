package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Estrutura padronizada para resposta
type Address struct {
	API        string
	CEP        string
	Logradouro string
	Bairro     string
	Cidade     string
	UF         string
	Err        error
}

// BrasilAPI response
type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

// ViaCEP response
type ViaCEPResponse struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
}

// -------------------- REQUESTS --------------------

func fetchBrasilAPI(ctx context.Context, cep string, ch chan<- Address) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- Address{Err: err}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Address{Err: err}
		return
	}
	defer resp.Body.Close()

	var data BrasilAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		ch <- Address{Err: err}
		return
	}

	ch <- Address{
		API:        "BrasilAPI",
		CEP:        data.Cep,
		Logradouro: data.Street,
		Bairro:     data.Neighborhood,
		Cidade:     data.City,
		UF:         data.State,
	}
}

func fetchViaCEP(ctx context.Context, cep string, ch chan<- Address) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- Address{Err: err}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Address{Err: err}
		return
	}
	defer resp.Body.Close()

	var data ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		ch <- Address{Err: err}
		return
	}

	ch <- Address{
		API:        "ViaCEP",
		CEP:        data.Cep,
		Logradouro: data.Logradouro,
		Bairro:     data.Bairro,
		Cidade:     data.Localidade,
		UF:         data.Uf,
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <CEP>")
		return
	}

	cep := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch := make(chan Address)

	// Goroutines paralelas
	go fetchBrasilAPI(ctx, cep, ch)
	go fetchViaCEP(ctx, cep, ch)

	select {
	case result := <-ch:
		if result.Err != nil {
			fmt.Println("Erro:", result.Err)
			return
		}

		fmt.Println("Endereço encontrado:")
		fmt.Println("API:", result.API)
		fmt.Println("CEP:", result.CEP)
		fmt.Println("Rua:", result.Logradouro)
		fmt.Println("Bairro:", result.Bairro)
		fmt.Println("Cidade:", result.Cidade)
		fmt.Println("Estado:", result.UF)

	case <-ctx.Done():
		fmt.Println("Timeout: nenhuma API respondeu em tempo hábil")
	}
}
