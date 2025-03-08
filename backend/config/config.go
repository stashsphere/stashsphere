package config

type StashsphereServeConfig struct {
	User     string  `koanf:"database.user"`
	Name     string  `koanf:"database.name"`
	Host     string  `koanf:"database.host"`
	Password *string `koanf:"database.password"`
	Port     *uint16 `koanf:"database.port"`
	SslMode  *string `koanf:"database.sslmode"`

	ListenAddress string `koanf:"listenAddress"`

	PrivateKey     string `koanf:"auth.privateKey"`
	ImagePath      string `koanf:"imagePath"`
	ImageCachePath string `koanf:"imageCachePath"`

	InviteEnabled bool   `koanf:"invites.enabled"`
	InviteCode    string `koanf:"invites.code"`

	AllowedDomains []string `koanf:"domains.allowed"`
	ApiDomain      string   `koanf:"domains.api"`
}

type StashSphereMigrateConfig struct {
	User     string  `koanf:"database.user"`
	Name     string  `koanf:"database.name"`
	Host     string  `koanf:"database.host"`
	Password *string `koanf:"database.password"`
	Port     *uint16 `koanf:"database.port"`
	SslMode  *string `koanf:"database.sslmode"`
}
