package workers

import (
	"context"
	"log"
	"sync"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/jackc/pgx/v5"
)

type DocumentProcessor struct {
	ctx       context.Context
	store     db.Store
	converter utils.Converter
	storage   utils.Storage
}

func NewDocumentProcessor(store db.Store, ctx context.Context, converter utils.Converter, storage utils.Storage) *DocumentProcessor {
	return &DocumentProcessor{
		store:     store,
		ctx:       ctx,
		converter: converter,
		storage:   storage,
	}
}

func (dp *DocumentProcessor) Work(errChan chan error) {
	for {
		// Check for shutdown signal
		select {
		case <-dp.ctx.Done():
			// Graceful shutdown
			return
		default:
		}

		// Fetch entries to process
		entries, err := dp.store.GetEntriesByStatus(dp.ctx, db.GetEntriesByStatusParams{
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

	documents, err := dp.store.GetDocumentsByEntryId(dp.ctx, db.GetDocumentsByEntryIdParams{
		EntryID: entry.ID,
		Limit:   10,
	})

	if err != nil {
		log.Printf("error reading entry with id %s ", entry.ID)
	}

	files, err := dp.storage.GetFiles(entry.ID.String())
	var status string

	if err != nil {
		done <- false
		status = "failed"
		return
	}
	doneChan := make(chan bool)

	if entry.Operation == "merge" {
		go dp.converter.Merge(files, entry.ID.String(), doneChan)
	} else {
		go dp.converter.Convert(files, entry.ID.String(), doneChan)
	}
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
			filename, err := dp.storage.TransformName(entry.ID.String(), documents[i].Filename)
			pages, _ := dp.converter.GetPageCount(filename)
			if err != nil {
				pages = 0
			}
			dp.store.UpdatePageCount(dp.ctx, db.UpdatePageCountParams{
				PageCount: pages,
				ID:        documents[i].ID,
			})
		}(documents, i)
	}
	wg.Wait()

	dp.store.UpdateStatus(dp.ctx, db.UpdateStatusParams{
		Status: status,
		ID:     entry.ID,
	})
	done <- true
}
