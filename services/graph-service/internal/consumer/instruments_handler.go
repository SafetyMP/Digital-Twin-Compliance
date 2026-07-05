package consumer

import (
	"context"

	"github.com/digital-twin/platform/services/graph-service/internal/debezium"
	"github.com/digital-twin/platform/services/graph-service/internal/graph"
)

func (h *Handler) HandleInstrumentMessage(ctx context.Context, data []byte) error {
	row, err := debezium.ParseInstrumentMessage(data)
	if err != nil || row == nil {
		return err
	}
	owner := debezium.StringField(row, "owner_entity_id")
	counterparty := debezium.StringField(row, "counterparty_id")
	if owner == "" || counterparty == "" {
		return nil
	}
	instrumentID := debezium.StringField(row, "instrument_id")
	if instrumentID == "" {
		return nil
	}
	return h.store.UpsertExposure(ctx, graph.ExposureInput{
		FromEntityID: owner,
		ToEntityID:   counterparty,
		EdgeKey:      instrumentID,
		ExposureType: debezium.StringField(row, "instrument_type"),
		NotionalEur:  debezium.FloatField(row, "notional_amount"),
		Layer:        classifyLayer(debezium.StringField(row, "instrument_type")),
		InstrumentID: instrumentID,
	})
}
