package postgres

// ClaimQuery returns the claim statement of repo, for the integration tests' query plan checks.
func ClaimQuery(repo *Repository) string {
	return repo.q.claim
}
