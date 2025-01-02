package workers

import (
	"context"
	"log"

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
			go dp.processEntry(entries[i], complete[i])
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
		done <- false
		return
	}

	files, err := dp.storage.GetFiles(entry.ID.String())

	if err != nil {
		done <- false
		return
	}
	doneChan := make(chan bool)

	if entry.Operation == "merge" {
		go dp.converter.Merge(files, entry.ID.String(), doneChan)
	} else {
		go dp.converter.Convert(files, entry.ID.String(), doneChan)
	}
	ok := <-doneChan

	if !ok {
		dp.updateOnCompletion(entry, "failed")
	} else {
		doneChans := make([]chan bool, len(documents))
		for i := range documents {
			doneChans[i] = make(chan bool)
			go dp.updateDocumentPages(entry, &documents[i], doneChans[i])
		}
		for i := range doneChans {
			<-doneChans[i]
		}

		dp.updateOnCompletion(entry, "success")

	}
	done <- ok
}

func (dp *DocumentProcessor) updateOnCompletion(entry db.Entry, status string) {
	dp.store.UpdateStatus(dp.ctx, db.UpdateStatusParams{
		Status: status,
		ID:     entry.ID,
	})
}

func (dp *DocumentProcessor) updateDocumentPages(entry db.Entry, document *db.Document, done chan bool) {
	filename, _ := dp.storage.TransformName(entry.ID.String(), document.Filename)
	pages, err := dp.converter.GetPageCount(filename)
	if err != nil {
		done <- false
		return
	}
	dp.store.UpdatePageCount(dp.ctx, db.UpdatePageCountParams{
		PageCount: pages,
		ID:        document.ID,
	})
	done <- true
}
