package tinyweb

type Config struct {
	EnableGzip bool
	SecretKey  string
	Logger     Logger
	paramMap   map[string]interface{}
}

var WebConfig Config

func init() {
	WebConfig.SecretKey = "asdfsadfwexczv asfwe"
	WebConfig.EnableGzip = true
	WebConfig.Logger = &DefaultLogger{}
	WebConfig.paramMap = make(map[string]interface{})
}

func (c *Config) SetParam(key string, val interface{}) {
	c.paramMap[key] = val
}

func (c *Config) GetParam(key string) (interface{}, bool) {
	if val, has := c.paramMap[key]; has {
		return val, true
	} else {
		return nil, false
	}
}
