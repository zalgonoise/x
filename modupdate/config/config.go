package config

type Config struct {
	CronSchedule string
	Repository   Repository
	Checkout     Checkout
	Update       Update
	Push         Push
}

type Repository struct {
	Path       string
	ModulePath string
	Branch     string
	Username   string
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
	DryRun           bool
	CommitMessage    string
	CommandOverrides []string
	FilesOverride    []string
}
