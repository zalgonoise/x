package mysql

import "github.com/zalgonoise/x/gacha"

// ClaimQuery returns the locking read of repo's claims, for the integration tests' query plan checks.
func ClaimQuery(repo *Repository) string {
	return repo.claimQuery
}

// SettleCompletedQuery returns the statement settling n completed outcomes, for the query plan checks.
func SettleCompletedQuery(repo *Repository, outcomes []gacha.Outcome) (string, []any) {
	stmt := repo.settleCompleted(outcomes)

	return stmt.query, stmt.args
}
