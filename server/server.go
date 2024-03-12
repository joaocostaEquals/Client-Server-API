package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	Valor string `json:"bid"`
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "cotacao.db")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Conexão com o banco de dados estabelecida")
}

func insertCotacaoBD(ctx context.Context, cotacao *Cotacao) error {
	//contexto de 10ms
	ctx, cancelDB := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancelDB()
	query := "INSERT INTO cotacao (bid) VALUES (?)"
	_, err := db.ExecContext(ctx, query, cotacao.Valor)
	return err
}

func main() {
	http.HandleFunc("/cotacao", dolarHandler)
	log.Printf("Servidor escutando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func dolarHandler(w http.ResponseWriter, r *http.Request) {
	cotacao, err := getCotacao()
	if err != nil {
		log.Printf("Erro ao buscar cotação do dolar: %v", err)
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	err = insertCotacaoBD(r.Context(), cotacao)
	if err != nil {
		log.Printf("Erro ao inserir cotação no banco de dados: %v", err)
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotacao)
}

func getCotacao() (*Cotacao, error) {
	apiURL := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]Cotacao
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	cotacao := data["USDBRL"]
	return &cotacao, nil
}
