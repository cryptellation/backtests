//go:build integration
// +build integration

package sql

import (
	"context"
	"testing"

	"github.com/cryptellation/backtests/configs"
	"github.com/cryptellation/backtests/configs/sql/down"
	"github.com/cryptellation/backtests/configs/sql/up"
	"github.com/cryptellation/backtests/svc/db"
	"github.com/cryptellation/dbmigrator"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

func TestBacktestSuite(t *testing.T) {
	suite.Run(t, new(BacktestSuite))
}

type BacktestSuite struct {
	db.BacktestSuite
}

func (suite *BacktestSuite) SetupSuite() {
	db, err := New(context.Background(), viper.GetString(configs.EnvSQLDSN))
	suite.Require().NoError(err)

	mig, err := dbmigrator.NewMigrator(context.Background(), db.db, up.Migrations, down.Migrations, nil)
	suite.Require().NoError(err)
	suite.Require().NoError(mig.MigrateToLatest(context.Background()))

	suite.DB = db
}

func (suite *BacktestSuite) SetupTest() {
	db := suite.DB.(*Activities)
	suite.Require().NoError(db.Reset(context.Background()))
}
