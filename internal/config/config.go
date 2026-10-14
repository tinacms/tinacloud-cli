package config

type CoreConfig struct {
	Dev bool
}

type TinaCloudConfig struct {
	Endpoint string
}

type Config struct {
	Core      CoreConfig
	TinaCloud TinaCloudConfig
}
