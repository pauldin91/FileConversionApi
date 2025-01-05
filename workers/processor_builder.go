package workers

import (
	"context"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
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

func (builder *ProcessorBuilder) WithStore(store db.Store) *ProcessorBuilder {
	builder.processor.store = store
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

func (builder *ProcessorBuilder) WithCtx(parentCtx context.Context) *ProcessorBuilder {
	ctx, cancel := context.WithCancel(parentCtx)
	builder.processor.ctx = ctx
	builder.processor.cancel = cancel
	return builder
}

func (builder *ProcessorBuilder) Build() DocumentProcessor {
	return builder.processor
}
