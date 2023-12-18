package gcommon

var (
	GameConf = &GameServerConf{}
)

type GameServerConf struct {
	Host string
	Port int

	MysqlAddr     string
	MysqlUser     string
	MysqlPassword string
	MysqlDb       string

	RedisGameCfgAddr string
	RedisGameCfgPass string

	HallHost   string
	HallPort   int
	HallSecret string

	GameHost string
	GamePort int
	LogPath  string
	LogLevel string
}
