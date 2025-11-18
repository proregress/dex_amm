package solmodel

import (
	"context"

	"richcode.cc/dex/pkg/constants"

	. "github.com/klen-ygs/gorm-zero/gormc/sql"
	"gorm.io/gorm"
)

// avoid unused err
var _ = InitField
var _ BlockModel = (*customBlockModel)(nil)

type (
	// BlockModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBlockModel.
	BlockModel interface {
		blockModel
		customBlockLogicModel
	}

	customBlockLogicModel interface {
		WithSession(tx *gorm.DB) BlockModel
	}

	customBlockModel struct {
		*defaultBlockModel
	}
)

func (c customBlockModel) WithSession(tx *gorm.DB) BlockModel {
	newModel := *c.defaultBlockModel
	c.defaultBlockModel = &newModel
	c.conn = tx
	return c
}

// NewBlockModel returns a model for the database table.
func NewBlockModel(conn *gorm.DB) BlockModel {
	return &customBlockModel{
		defaultBlockModel: newBlockModel(conn),
	}
}

func (m *defaultBlockModel) customCacheKeys(data *Block) []string {
	if data == nil {
		return []string{}
	}
	return []string{}
}

func (m *defaultBlockModel) FindOneByNearSlot(ctx context.Context, slot int64) (*Block, error) {
	var resp Block
	err := m.conn.WithContext(ctx).Model(&Block{}).Where("`slot` < ? and `status` = ?", slot, constants.BlockProcessed).Order("slot desc").First(&resp).Error
	return &resp, err
}
