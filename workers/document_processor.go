package workers

import (
	"context"
	"log"
	"time"

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
		select {
		case <-dp.ctx.Done():
			return
		default:
		}
		entries, err := dp.store.GetEntriesByStatus(dp.ctx, db.GetEntriesByStatusParams{
			Status: string(utils.Processing),
			Limit:  10,
		})
		if err != nil {
			if err != pgx.ErrNoRows {
				errChan <- err
			}
			continue
		}
		complete := make([]chan bool, len(entries))
		for i := range entries {
			complete[i] = make(chan bool)
			go dp.processEntry(entries[i], complete[i])
		}

		for i := range complete {
			<-complete[i]
		}
		if len(entries) == 0 {
			time.Sleep(time.Minute)
		}
	}
}

func (dp *DocumentProcessor) processEntry(entry db.Entry, done chan bool) {
	documents, err := dp.store.GetDocumentsByEntryId(dp.ctx, db.GetDocumentsByEntryIdParams{
		EntryID: entry.ID,
		Limit:   10,
	})
	start := time.Now()
	if err != nil {
		log.Printf("error reading entry with id %s ", entry.ID)
		dp.updateRetries(entry, done)
		done <- false
		return
	}
	for i := range documents {
		filename := dp.storage.GetFilename(entry.ID.String(), documents[i].Filename)
		if !dp.storage.FileExists(filename) {
			dp.updateRetries(entry, done)
			done <- false
			return
		}
	}

	files, err := dp.storage.GetFiles(entry.ID.String())
	if err != nil || len(files) == 0 {
		dp.updateRetries(entry, done)
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

func (dp *DocumentProcessor) updateOnCompletion(entry db.Entry, status string, timeElapsed float64) {
	dp.store.UpdateProcessed(dp.ctx, db.UpdateProcessedParams{
		Status:      status,
		ID:          entry.ID,
		TimeElapsed: timeElapsed,
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

func (dp *DocumentProcessor) updateRetries(entry db.Entry, done chan bool) {
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
}
