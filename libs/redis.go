// hello project main.go
package libs

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

var RedisInfo *RedisFunc
var RedisLogName string = "redis"

type RedisFunc struct {
	Host     string
	Port     string
	Db       string
	Password string
	Pre      string
	Pool     *redis.Pool
}

func (r *RedisFunc) InitConf() {
	//初始化redis
	r.Pool = r.newPool(r.Host+":"+r.Port, r.Password, r.Db)
	//	fmt.Println(r.Pool)
}

func NewRedis(conf *S1) *RedisFunc {

	RedisInfo = &RedisFunc{conf.Redis.Host, conf.Redis.Port, conf.Redis.Db,
		conf.Redis.Password, conf.Redis.Pre, nil}
	//	fmt.Println(RedisInfo)
	RedisInfo.InitConf()
	return RedisInfo
}

func NewRedisConnect(host, port, password, db string) *RedisFunc {

	cc := &RedisFunc{}
	//fmt.Println(cc)
	cc.Pool = cc.newPool(host+":"+port, password, db)
	return cc
}

// 重写生成连接池方法
func (r *RedisFunc) newPool(server, password, db string) *redis.Pool {

	return &redis.Pool{
		MaxIdle:     10,
		MaxActive:   600, // max number of connections
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if db != "" {
				if _, err := c.Do("SELECT", db); err != nil {
					c.Close()
					return nil, err
				}
			}

			return c, err
		},
	}
}

//推送队列push
func (r *RedisFunc) Lpush(dbkey string, content string) (int, error) {
	c := r.Pool.Get()
	defer c.Close()
	v, err := redis.Int(c.Do("lpush", dbkey, content))
	if err != nil && err != redis.ErrNil {
		errorInfo := fmt.Sprintf("redis lpush error:", err.Error())
		Log.Error(RedisLogName, errorInfo)
	}
	return v, err
}

//从队列获取数据
func (r *RedisFunc) Rpop(dbkey string) (string, error) {
	c := r.Pool.Get()
	defer c.Close()
	v, err := redis.String(c.Do("rpop", dbkey))
	if err != nil && err != redis.ErrNil {
		errorInfo := fmt.Sprintf("redis rpop error:%s", err.Error())
		Log.Error(RedisLogName, errorInfo)
	}
	return v, err
}

//get
func (r *RedisFunc) Get(dbkey string) (string, error) {
	c := r.Pool.Get()
	defer c.Close()

	v, err := redis.String(c.Do("get", dbkey, ""))
	if err != nil && err != redis.ErrNil {
		errorInfo := fmt.Sprintf("redis rpop error:", err.Error())
		Log.Error(RedisLogName, errorInfo)
	}
	return v, err
}

//get
func (r *RedisFunc) Expire(dbkey string, t int) (string, error) {
	c := r.Pool.Get()
	defer c.Close()

	v, err := redis.String(c.Do("expire", dbkey, t))
	if err != nil && err != redis.ErrNil {
		errorInfo := fmt.Sprintf("redis expire error:", err.Error())
		Log.Error(RedisLogName, errorInfo)
	}
	return v, err
}

//增加内容到集合中
func (r *RedisFunc) Sadd(key string, content string) (int, error) {
	c := r.Pool.Get()
	defer c.Close()
	v, err := redis.Int(c.Do("sadd", key, content))
	if err != nil && err != redis.ErrNil {

		errorInfo := fmt.Sprintf("redis sadd error:", err.Error())
		Log.Error(RedisLogName, errorInfo)
	}
	return v, err
}

//
func (r *RedisFunc) Smembers(key string) []string {
	c := r.Pool.Get()
	defer c.Close()
	v, err := redis.Strings(c.Do("smembers", key, ""))
	if err != nil && err != redis.ErrNil {
		errorInfo := fmt.Sprintf("redis smembers error:", err)
		Log.Error(RedisLogName, errorInfo)
	}
	return v
}

//
func (r *RedisFunc) Spop(key string) string {
	c := r.Pool.Get()
	defer c.Close()
	v, err := redis.String(c.Do("spop", key))
	//fmt.Println("===========", v, err)
	if err != nil && err != redis.ErrNil {
		errorInfo := fmt.Sprintf("redis spop error:%s", err.Error())
		Log.Error(RedisLogName, errorInfo)
	}
	return v
}

//
func (r *RedisFunc) Llen(key string) (int, error) {
	c := r.Pool.Get()
	defer c.Close()
	v, err := redis.Int(c.Do("llen", key))
	if err != nil {
		errorInfo := fmt.Sprintf("redis llen error:%s", err)
		Log.Error(RedisLogName, errorInfo)
	}
	return v, err
}

//
func (r *RedisFunc) Del(key string) (int, error) {
	c := r.Pool.Get()
	defer c.Close()
	v, err := redis.Int(c.Do("del", key))
	if err != nil {
		errorInfo := fmt.Sprintf("redis del error:%s", err)
		Log.Error(RedisLogName, errorInfo)
	}
	return v, err
}

func (r *RedisFunc) Sunionstore(key, key2 string) (int, error) {
	c := r.Pool.Get()
	defer c.Close()
	v, err := redis.Int(c.Do("SUNIONSTORE", key, key2))
	if err != nil {
		errorInfo := fmt.Sprintf("redis SUNIONSTORE error:%s", err)
		Log.Error(RedisLogName, errorInfo)
	}
	return v, err
}

//
func (r *RedisFunc) Smove(key, otherkey, value string) []string {
	c := r.Pool.Get()
	defer c.Close()
	v, err := redis.Strings(c.Do("smove", key, otherkey, value))
	if err != nil && err != redis.ErrNil {
		errorInfo := fmt.Sprintf("redis smove error:", err)
		Log.Error(RedisLogName, errorInfo)
	}
	return v
}
