package transcribe

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/redis/go-redis/v9"
)

const (
	audioDir       = "audio"
	openAIEndpoint = "https://api.openai.com/v1/audio/transcriptions"
)

type Transcriber struct {
	openAIKey string

	db    *sql.DB
	redis *redis.Client
}

func NewTranscriber(db *sql.DB, redis *redis.Client, openAPIKey string) *Transcriber {
	log.Println("Creating new transcriber...")
	return &Transcriber{
		db:        db,
		redis:     redis,
		openAIKey: openAPIKey,
	}
}

func (t *Transcriber) AttemptTranscription(filename string) (string, error) {
	log.Println("Transcribing audio file:", filename)

	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("error opening file %s: %w", filename, err)
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("error creating form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("error copying file content: %w", err)
	}

	err = writer.WriteField("model", "whisper-1")
	if err != nil {
		return "", fmt.Errorf("error writing form field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("error closing multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", openAIEndpoint, &requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+t.openAIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON response: %w", err)
	}

	transcription, ok := result["text"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format: %v", result)
	}

	fmt.Printf("Transcription of %s: %s\n", filename, transcription)

	err = t.Save(filename, transcription)
	if err != nil {
		return "", fmt.Errorf("error saving transcription to database: %w", err)
	}

	return transcription, nil
}

func (t *Transcriber) GenerateHeadline(articleText string) (string, error) {
	// Prepare request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"prompt":     articleText,
		"max_tokens": 20, // Adjust based on desired headline length
	})
	if err != nil {
		return "", err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", openAIEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.openAIKey))

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse response
	// var result map[string]interface{}
	// if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
	// 	return "", err
	// }

	// Extract headline from response

	b := new(bytes.Buffer)
	b.ReadFrom(resp.Body)

	return b.String(), nil
}

func (t *Transcriber) Save(fileName, transcription string) error {
	query := `
	INSERT INTO transcriptions (file_name, transcription)
	VALUES ($1, $2)
	`
	_, err := t.db.Exec(query, fileName, transcription)
	if err != nil {
		return fmt.Errorf("error inserting transcription into database: %w", err)
	}
	return nil
}
