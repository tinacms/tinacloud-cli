package config

type CoreConfig struct {
	Dev bool
}

type TinaCloudConfig struct {
	IdentityEndpoint string
	ContentEndpoint  string
}

type Config struct {
	Core      CoreConfig
	TinaCloud TinaCloudConfig
}
