package shadowsocks

type Key struct {
	Id     string `yaml:"id"`
	Port   int    `yaml:"port"`
	Cipher string `yaml:"cipher"`
	Secret string `yaml:"secret"`
}
