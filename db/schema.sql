CREATE TABLE IF NOT EXISTS transcriptions (
    id SERIAL PRIMARY KEY,
    file_name TEXT NOT NULL,
    transcription TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
