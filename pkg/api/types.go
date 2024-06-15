package api

type Article struct {
	ID            int    `json:"id"`
	FileName      string `json:"file_name"`
	Transcription string `json:"transcription"`
}
