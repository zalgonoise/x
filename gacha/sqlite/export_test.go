package sqlite

import "github.com/zalgonoise/x/gacha"

// ClaimQuery returns the read of repo's claims, for the query plan checks.
func ClaimQuery(repo *Repository) string {
	return repo.claimQuery
}

// ViewClaimQuery returns the read of a view's claims, for the query plan checks.
func ViewClaimQuery(source gacha.Source[[]byte]) string {
	return source.(view).claimQuery()
}

// SettleCompletedQuery returns the statement settling completed outcomes, for the query plan checks.
func SettleCompletedQuery(repo *Repository, outcomes []gacha.Outcome) (string, []any) {
	stmt := repo.settleCompleted(outcomes)

	return stmt.query, stmt.args
}
