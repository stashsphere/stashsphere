package config

type StashSphereDatabaseConfig struct {
	User     string  `koanf:"user"`
	Name     string  `koanf:"name"`
	Host     string  `koanf:"host"`
	Password *string `koanf:"password"`
	Port     *uint16 `koanf:"port"`
	SslMode  *string `koanf:"sslmode"`
}

type StashSphereServeConfig struct {
	Database StashSphereDatabaseConfig `koanf:"database"`

	ListenAddress string `koanf:"listenAddress"`

	Auth struct {
		PrivateKey string `koanf:"privateKey"`
	} `koanf:"auth"`

	Image struct {
		Path      string `koanf:"path"`
		CachePath string `koanf:"cachePath"`
	} `koanf:"image"`

	Invites struct {
		Enabled    bool   `koanf:"enabled"`
		InviteCode string `koanf:"code"`
	} `koanf:"invites"`

	Domains struct {
		AllowedDomains []string `koanf:"allowed"`
		ApiDomain      string   `koanf:"api"`
	} `koanf:"domains"`

	FrontendUrl  string `koanf:"frontendUrl"`
	InstanceName string `koanf:"instanceName"`

	Email StashSphereMailConfig `koanf:"email"`
}

type StashSphereMigrateConfig struct {
	Database StashSphereDatabaseConfig `koanf:"database"`
}

type StashSphereMailConfig struct {
	Backend  string `koanf:"backend"`
	FromAddr string `koanf:"fromAddr"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	Host     string `koanf:"host"`
	Port     uint16 `koanf:"port"`
}
