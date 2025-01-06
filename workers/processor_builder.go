package workers

import (
	"context"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Generic ProcessorBuilder
type ProcessorBuilder struct {
	processor *PdfProcessor
}

func Builder() *ProcessorBuilder {
	return &ProcessorBuilder{
		processor: &PdfProcessor{},
	}
}

func (builder *ProcessorBuilder) WithStore(pool *pgxpool.Pool) *ProcessorBuilder {
	processorCtx, processorCancel := context.WithCancel(context.Background()) // Create context for the processor
	processorStore := db.NewStore(pool)

	builder.processor.ctx = processorCtx
	builder.processor.cancel = processorCancel
	builder.processor.store = processorStore
	builder.processor.pool = pool
	return builder
}

func (builder *ProcessorBuilder) WithStorage(storage utils.Storage) *ProcessorBuilder {
	builder.processor.storage = storage
	return builder
}

func (builder *ProcessorBuilder) WithConverter(converter utils.Converter) *ProcessorBuilder {
	builder.processor.converter = converter
	return builder
}

func (builder *ProcessorBuilder) Build() DocumentProcessor {
	return builder.processor
}
