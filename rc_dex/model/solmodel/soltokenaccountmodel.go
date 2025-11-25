package solmodel

import (
	. "github.com/klen-ygs/gorm-zero/gormc/sql"
	"gorm.io/gorm"
)

// avoid unused err
var _ = InitField
var _ SolTokenAccountModel = (*customSolTokenAccountModel)(nil)

type (
	// SolTokenAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSolTokenAccountModel.
	SolTokenAccountModel interface {
		solTokenAccountModel
		customSolTokenAccountLogicModel
	}

	customSolTokenAccountLogicModel interface {
		WithSession(tx *gorm.DB) SolTokenAccountModel
	}

	customSolTokenAccountModel struct {
		*defaultSolTokenAccountModel
	}
)

func (c customSolTokenAccountModel) WithSession(tx *gorm.DB) SolTokenAccountModel {
	newModel := *c.defaultSolTokenAccountModel
	c.defaultSolTokenAccountModel = &newModel
	c.conn = tx
	return c
}

// NewSolTokenAccountModel returns a model for the database table.
func NewSolTokenAccountModel(conn *gorm.DB) SolTokenAccountModel {
	return &customSolTokenAccountModel{
		defaultSolTokenAccountModel: newSolTokenAccountModel(conn),
	}
}

func (m *defaultSolTokenAccountModel) customCacheKeys(data *SolTokenAccount) []string {
	if data == nil {
		return []string{}
	}
	return []string{}
}
