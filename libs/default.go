package libs

var (
	CONF  map[string]*S1
	REDIS map[string]*RedisFunc
	LOG   map[string]*Loger
)

//多配置文件使用
func InitInfo(info map[string]string) {

	count := len(info)
	CONF = make(map[string]*S1, count)
	REDIS = make(map[string]*RedisFunc, count)
	LOG = make(map[string]*Loger, count)

	for k, v := range info {
		CONF[k] = StartConf(v)
		LOG[k] = NewLoger(CONF[k])
		REDIS[k] = NewRedis(CONF[k])
	}

}

//单个配置文件使用
func InitSingleInfo(path string) {
	ServerConf = StartConf(path)
	Log = NewLoger(ServerConf)
	RedisInfo = NewRedis(ServerConf)
}
