package config

type StashsphereServeConfig struct {
	User     string `koanf:"database.user"`
	Password string `koanf:"database.password"`
	Name     string `koanf:"database.name"`
	Host     string `koanf:"database.host"`

	ListenAddress string `koanf:"listenAddress"`

	PrivateKey string `koanf:"auth.privateKey"`
	ImagePath  string `koanf:"imagePath"`

	InviteEnabled bool   `koanf:"invites.enabled"`
	InviteCode    string `koanf:"invites.code"`
}

type StashSphereMigrateConfig struct {
	User     string `koanf:"database.user"`
	Password string `koanf:"database.password"`
	Name     string `koanf:"database.name"`
	Host     string `koanf:"database.host"`
}
