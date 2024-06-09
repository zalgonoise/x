package config

type Config struct {
	Repository Repository
	Checkout   Checkout
	Update     Update
	Push       Push
}

type Repository struct {
	Path       string
	ModulePath string
	Token      string
}

type Checkout struct {
	Persist         bool
	Path            string
	GitPath         string
	CommandOverride string
}

type Update struct {
	GoBin               string
	GitCommandOverrides []string
	GoCommandOverrides  []string
}

type Push struct {
	DryRun          bool
	CommandOverride string
	CommitMessage   string
	FilesOverride   []string
}
