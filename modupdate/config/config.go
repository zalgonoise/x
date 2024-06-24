package config

type Config struct {
	DatabaseURI string  `json:"database_uri,omitempty"`
	Events      *Events `json:"events,omitempty"`
	Tasks       []*Task `json:"tasks,omitempty"`
}

type Task struct {
	CronSchedule string     `json:"cron_schedule,omitempty"`
	Repository   Repository `json:"repository"`
	Checkout     Checkout   `json:"checkout"`
	Update       Update     `json:"update"`
	Check        Check      `json:"check"`
	Push         Push       `json:"push"`
}

type Repository struct {
	Path       string `json:"uri,omitempty"`
	ModulePath string `json:"module,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Username   string `json:"username,omitempty"`
	Token      string `json:"token,omitempty"`
}

type Events struct {
	DiscordToken string `json:"discord_token,omitempty"`
	Skip         bool   `json:"skip,omitempty"`
	BufferSize   int    `json:"buffer_size,omitempty"`
}

type Checkout struct {
	Persist          bool     `json:"persist,omitempty"`
	Path             string   `json:"path,omitempty"`
	GitPath          string   `json:"git_path,omitempty"`
	CommandOverrides []string `json:"command_overrides,omitempty"`
}

type Update struct {
	GoBin               string   `json:"go_bin,omitempty"`
	GitCommandOverrides []string `json:"git_command_overrides,omitempty"`
	GoCommandOverrides  []string `json:"go_command_overrides,omitempty"`
}

type Check struct {
	Skip             bool     `json:"skip"`
	CommandOverrides []string `json:"command_overrides,omitempty"`
}

type Push struct {
	DryRun           bool     `json:"dry_run,omitempty"`
	CommitMessage    string   `json:"commit_message,omitempty"`
	CommandOverrides []string `json:"command_overrides,omitempty"`
	FilesOverride    []string `json:"files_override,omitempty"`
}
