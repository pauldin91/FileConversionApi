package workers

import (
	"context"
	"log"
	"time"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/jackc/pgx/v5"
)

type PdfProcessor struct {
	ctx       context.Context
	store     db.Store
	converter utils.Converter
	storage   utils.Storage
	cancel    context.CancelFunc
}

func (dp *PdfProcessor) Work() error {
	for {
		select {
		case <-dp.ctx.Done():
			return nil
		default:
			// Use a separate goroutine to avoid blocking here
			entries, err := dp.store.GetEntriesByStatus(dp.ctx, db.GetEntriesByStatusParams{
				Status: string(utils.Processing),
				Limit:  10,
			})
			if err != nil {
				if err == pgx.ErrNoRows {
					continue
				} else {
					return err
				}
			}

			complete := make([]chan bool, len(entries))
			for i := range entries {
				complete[i] = make(chan bool)
				go dp.processEntry(entries[i], complete[i]) // Ensure processing happens in the background
			}

			// Wait for all goroutines to complete
			for i := range complete {
				<-complete[i]
			}
		}
	}
}

func (dp *PdfProcessor) processEntry(entry db.Entry, done chan bool) {
	documents, err := dp.store.GetDocumentsByEntryId(dp.ctx, db.GetDocumentsByEntryIdParams{
		EntryID: entry.ID,
		Limit:   10,
	})
	start := time.Now()
	if err != nil {
		log.Printf("error reading entry with id %s ", entry.ID)
		dp.updateRetries(entry, done)
		return
	}
	var files []string
	for i := range documents {
		filename := dp.storage.GetFilename(entry.ID.String(), documents[i].Filename)
		if !dp.storage.FileExists(filename) {
			dp.updateRetries(entry, done)
			return
		}
		files = append(files, filename)
	}

	if len(files) == 0 || len(documents) != len(files) {
		dp.updateRetries(entry, done)
		return
	}

	doneChan := make(chan bool)

	if entry.Operation == string(utils.Merge) {
		go dp.converter.Merge(files, entry.ID.String(), doneChan)
	} else {
		go dp.converter.Convert(files, entry.ID.String(), doneChan)
	}
	ok := <-doneChan

	if !ok {
		dp.updateOnCompletion(entry, string(utils.Failed), 0.0)
	} else {
		doneChans := make([]chan bool, len(documents))
		for i := range documents {
			doneChans[i] = make(chan bool)
			go dp.updateDocumentPages(entry, &documents[i], doneChans[i])
		}
		for i := range doneChans {
			<-doneChans[i]
		}
		timeElapsed := time.Since(start).Seconds()
		dp.updateOnCompletion(entry, string(utils.Success), timeElapsed)

	}
	done <- ok
}

func (dp *PdfProcessor) updateOnCompletion(entry db.Entry, status string, timeElapsed float64) {
	dp.store.UpdateProcessed(dp.ctx, db.UpdateProcessedParams{
		Status:      status,
		ID:          entry.ID,
		TimeElapsed: timeElapsed,
	})
}

func (dp *PdfProcessor) updateDocumentPages(entry db.Entry, document *db.Document, done chan bool) {
	filename, _ := dp.storage.GetConvertedFilename(entry.ID.String(), document.Filename)
	pages, err := dp.converter.GetPageCount(filename)
	if err != nil {
		done <- false
		return
	}
	dp.store.UpdatePageCount(dp.ctx, db.UpdatePageCountParams{
		PageCount: pages,
		ID:        document.ID,
	})
}

func (dp *PdfProcessor) updateRetries(entry db.Entry, done chan bool) {
	current, err := dp.store.GetEntry(dp.ctx, entry.ID)
	if err != nil {
		done <- false
		return
	}
	if current.MaxRetries == 0 {
		dp.updateOnCompletion(entry, string(utils.Failed), 0.0)
		done <- false
		return
	}
	dp.store.UpdateRetries(dp.ctx, db.UpdateRetriesParams{
		ID:         entry.ID,
		MaxRetries: current.MaxRetries - 1,
	})
	done <- true
}
