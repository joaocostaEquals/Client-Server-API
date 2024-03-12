package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Valor string `json:"bid"`
}

func main() {
	cotacao, err := getCotacao()
	if err != nil {
		log.Printf("Erro ao buscar cotação do dolar: %v", err)
		return
	}
	err = salvarArquivo("cotacao.txt", cotacao)
	if err != nil {
		log.Fatal("Erro ao salvar o resultado em um arquivo:", err)
	}
}

func getCotacao() (*Cotacao, error) {
	apiURL := "http://localhost:8080/cotacao"
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}

	response, err := client.Do(req)
	if err != nil {
		log.Println("Erro ao buscar cotação do dolar:", err.Error())
		return nil, err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		return nil, err
	}
	return &cotacao, nil
}

func salvarArquivo(nomeArquivo string, cotacao *Cotacao) error {
	// Abre o arquivo para escrita, cria se não existir
	file, err := os.OpenFile(nomeArquivo, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte("Dólar: " + cotacao.Valor + "\n"))
	if err != nil {
		return err
	}

	log.Printf("Resultado salvo em %s\n", nomeArquivo)
	return nil
}
