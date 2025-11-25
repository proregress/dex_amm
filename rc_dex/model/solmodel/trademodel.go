package solmodel

import (
	. "github.com/klen-ygs/gorm-zero/gormc/sql"
	"gorm.io/gorm"
)

// avoid unused err
var _ = InitField
var _ TradeModel = (*customTradeModel)(nil)

type (
	// TradeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTradeModel.
	TradeModel interface {
		tradeModel
		customTradeLogicModel
	}

	customTradeLogicModel interface {
		WithSession(tx *gorm.DB) TradeModel
	}

	customTradeModel struct {
		*defaultTradeModel
	}
)

func (c customTradeModel) WithSession(tx *gorm.DB) TradeModel {
	newModel := *c.defaultTradeModel
	c.defaultTradeModel = &newModel
	c.conn = tx
	return c
}

// NewTradeModel returns a model for the database table.
func NewTradeModel(conn *gorm.DB) TradeModel {
	return &customTradeModel{
		defaultTradeModel: newTradeModel(conn),
	}
}

func (m *defaultTradeModel) customCacheKeys(data *Trade) []string {
	if data == nil {
		return []string{}
	}
	return []string{}
}
