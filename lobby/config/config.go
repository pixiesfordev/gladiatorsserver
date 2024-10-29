package config

var (
	self = Config{}
)

type Config struct {
	env string // Env 版本
}

func Set(env string) {
	self = Config{
		env: env,
	}
}

func Conf() Config {
	return self
}

func Env() string {
	return self.env
}
