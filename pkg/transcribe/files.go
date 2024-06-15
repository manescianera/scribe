package transcribe

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/redis/go-redis/v9"
)

const hashKeyPrefix = "processed_files:"

func (t *Transcriber) Watch(ctx context.Context, directory string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("Event:", event)

				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("New file detected:", event.Name)
					go t.HandleNewFile(ctx, event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add(directory)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func (t *Transcriber) HandleNewFile(ctx context.Context, filePath string) {
	fmt.Println("Processing new file:", filePath)

	if isAudioFile(filePath) {
		fmt.Println("File is an audio file:", filePath)

		duplicate, err := isDuplicate(ctx, t.redis, hashKeyPrefix, filePath)
		if err != nil {
			log.Printf("Error checking if file is duplicate: %v\n", err)
			return
		}

		if duplicate {
			fmt.Println("File is a duplicate, skipping:", filePath)
			return
		}

		_, err = t.AttemptTranscription(filePath)
		if err != nil {
			log.Printf("Error transcribing file: %v\n", err)
		}

		// headline, err := generateHeadline(text)
		// if err != nil {
		// 	log.Fatal("Error generating headline:", err)
		// }
		// log.Println("Generated headline:", headline)

		// category, err := generateCategory(text)
		// if err != nil {
		// 	log.Fatal("Error generating category:", err)
		// }
		// log.Println("Generated category:", category)

	} else {
		fmt.Println("File is not an audio file, skipping:", filePath)
	}
}

func isAudioFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".mp3" || ext == ".wav"
}

func computeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func isDuplicate(ctx context.Context, redisClient *redis.Client, hashKeyPrefix, filePath string) (bool, error) {
	log.Println("Checking if file is a duplicate:", filePath)
	fileHash, err := computeFileHash(filePath)
	if err != nil {
		return false, err
	}

	exists, err := redisClient.Exists(ctx, hashKeyPrefix+fileHash).Result()
	if err != nil {
		return false, err
	}

	if exists > 0 {
		return true, nil
	}

	err = redisClient.Set(ctx, hashKeyPrefix+fileHash, true, 0).Err()
	if err != nil {
		return false, err
	}

	return false, nil
}
