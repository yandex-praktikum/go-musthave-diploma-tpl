package worker

type Config struct {
	Enabled               bool   `yaml:"Enabled"`
	AccrualSystemAddress  string `yaml:"AccrualSystemAddress"`
	CheckNewOrderInterval int    `yaml:"CheckInterval"`
	RateLimit             int    `yaml:"RateLimit"`
}
