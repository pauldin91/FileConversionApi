package workers

import (
	"context"
	"log"
	"os"
	"path"
	"sync"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/jackc/pgx/v5"
)

type DocumentProcessor struct {
	store     db.Store
	converter utils.Converter
}

func (dp *DocumentProcessor) Work(ctx context.Context, errChan chan error) {
	for {
		// Check for shutdown signal
		select {
		case <-ctx.Done():
			// Graceful shutdown
			return
		default:
		}

		// Fetch entries to process
		entries, err := dp.store.GetEntriesByStatus(context.Background(), db.GetEntriesByStatusParams{
			Status: "processing",
			Limit:  10,
		})
		if err != nil {
			if err != pgx.ErrNoRows {
				errChan <- err
			}
			continue
		}

		// Create channels to track processing completion
		complete := make([]chan bool, len(entries))
		for i := range entries {
			complete[i] = make(chan bool)
			go func(entry db.Entry, done chan bool) {
				defer close(done) // Ensure the channel is closed when processing is done
				dp.processEntry(entry, done)
			}(entries[i], complete[i])
		}

		// Wait for all entries to be processed
		for i := range complete {
			<-complete[i]
		}
	}
}

func (dp *DocumentProcessor) processEntry(entry db.Entry, done chan bool) {

	documents, err := dp.store.GetDocumentsByEntryId(context.Background(), db.GetDocumentsByEntryIdParams{
		EntryID: entry.ID,
	})

	if err != nil {
		log.Printf("error reading entry with id %s ", entry.ID)
	}

	files := make(map[string][]byte)
	for i := range documents {
		index := documents[i].Filename
		fullName := path.Join(entry.ID.String(), documents[i].Filename)
		contents, err := os.ReadFile(fullName)
		if err != nil {
			continue
		}
		files[index] = contents
	}
	doneChan := make(chan bool)

	if entry.Operation == "merge" {
		go dp.converter.Merge(files, entry.ID, doneChan)
	} else {
		go dp.converter.Convert(files, entry.ID, doneChan)
	}
	var status string
	if ok := <-doneChan; !ok {
		done <- ok
		status = "failed"
	} else {
		done <- true
		status = "success"
	}
	var wg sync.WaitGroup
	for i := range documents {
		wg.Add(1)
		go func(documents []db.Document, i int) {
			defer wg.Done()
			pages, _ := dp.converter.GetPageCount(documents[i].Filename, entry.ID)
			if err != nil {
				pages = 0
			}
			dp.store.UpdatePageCount(context.Background(), db.UpdatePageCountParams{
				PageCount: pages,
				ID:        documents[i].ID,
			})
		}(documents, i)
	}
	wg.Wait()

	dp.store.UpdateStatus(context.Background(), db.UpdateStatusParams{
		Status: status,
		ID:     entry.ID,
	})
	done <- true
}
