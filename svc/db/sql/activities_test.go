//go:build integration
// +build integration

package sql

import (
	"context"
	"testing"

	"github.com/cenkalti/backoff/v5"
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
	db, err := createTestDBClient(context.Background())
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

// createTestDBClient tries to create a new Activities client with backoff retry logic.
func createTestDBClient(ctx context.Context) (*Activities, error) {
	callback := func() (*Activities, error) {
		return New(ctx, viper.GetString(configs.EnvSQLDSN))
	}
	return backoff.Retry(ctx, callback,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(10),
	)
}
