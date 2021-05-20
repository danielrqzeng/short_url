package db

import (
	"context"
	"fmt"
	"iyfiysi.com/short_url/internal/pkg/logger"
	"iyfiysi.com/short_url/internal/pkg/utils"
	"strings"
	"time"

	"github.com/go-redsync/redsync"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
)

var pool *redigo.Pool

//RedisInit call to init redis pool when config ready
func RedisInit() {
	redisHost := viper.GetString("redis.host")
	redisPort := viper.GetString("redis.port")
	redisPass := viper.GetString("redis.pass")
	redisDB := viper.GetString("redis.db")
	if redisHost == "" || redisPort == "" {
		return
	}

	poolSize := 3
	pool = &redigo.Pool{
		MaxIdle:     poolSize,
		IdleTimeout: 240 * time.Second,
		//Dial:        func() (redigo.Conn, error) { return redigo.Dial("tcp", redisHost+":"+redisPort) },
		Dial: func() (conn redigo.Conn, e error) {
			conn, err := redigo.Dial("tcp", redisHost+":"+redisPort)
			if err != nil {
				return nil, err
			}
			if redisPass != "" {
				if _, err := conn.Do("AUTH", redisPass); err != nil {
					err = conn.Close()
					return nil, err
				}
			}
			if _, err := conn.Do("SELECT", redisDB); err != nil {
				err = conn.Close()
				return nil, err
			}
			return conn, nil
		},
		/*		TestOnBorrow: func(conn redigo.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				_, err := conn.Do("PING")
				return err
			},*/
	}
}

//Conn get redis connection from pool
func Conn() redigo.Conn {
	return pool.Get()
}

//RedisSet kv operator
func RedisSet(key string, val interface{}) error {
	c := Conn()
	defer c.Close()
	_, err := c.Do("SET", key, val)
	return err
}

//RedisSetEx ttl:time to live(second)
func RedisSetEx(key string, ttl int, val interface{}) error {
	c := Conn()
	defer c.Close()
	_, err := c.Do("SETEX", key, ttl, val)
	return err
}

//RedisExist key exist?
func RedisExist(key string) (bool, error) {
	c := Conn()
	defer c.Close()
	exist, err := redigo.Bool(c.Do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return exist, nil
}

//RedisGet get by key,for value is string
func RedisGet(key string) string {
	c := Conn()
	defer c.Close()
	val, err := redigo.String(c.Do("GET", key))
	if err != nil {
		if err == redigo.ErrNil {
			return ""
		}
		return ""
	}
	return val
}

//RedisGetInt get by key,for value is int
func RedisGetInt(key string) (uint64, error) {
	c := Conn()
	defer c.Close()
	val, err := redigo.Uint64(c.Do("GET", key))
	if err != nil {
		if err == redigo.ErrNil {
			return 0, err
		}
		return 0, err
	}
	return val, nil
}

//RedisGetBool get by key,for value is bool
func RedisGetBool(key string) bool {
	c := Conn()
	defer c.Close()
	val, err := redigo.Bool(c.Do("GET", key))
	if err != nil {
		if err == redigo.ErrNil {
			return false
		}
		panic(err)
	}
	return val
}

//RedisDel delete key
func RedisDel(key string) (err error) {
	c := Conn()
	defer c.Close()
	_, err = c.Do("DEL", key)
	if err != nil {
		return
	}
	return
}

//RedisExpire set key ttl
func RedisExpire(key string, sec int) bool {
	c := Conn()
	defer c.Close()
	n, err := c.Do("EXPIRE", key, sec)
	if err != nil {
		return false
	}
	if n == int64(1) {
		return true
	}
	return false
}

//RedisPExpire set key ttl(in ms)
func RedisPExpire(key string, ms int) bool {
	c := Conn()
	defer c.Close()
	n, err := c.Do("PEXPIRE", key, ms)
	if err != nil {
		return false
	}
	if n == int64(1) {
		return true
	}
	return false
}

//RedisLpush left push to list
func RedisLpush(key, val string) (listLen int64, err error) {
	c := Conn()
	defer c.Close()
	reply, err := c.Do("LPUSH", key, val)
	if err != nil {
		return
	}
	listLen = reply.(int64)
	return
}

//RedisRpop right pop from list
//ret:
//	listVal:列表的值，若是无值，则返回空
//	err: 连接错误之类的错误（不包含空值）
func RedisRpop(key string) (listVal string, err error) {
	c := Conn()
	defer c.Close()
	reply, err := redigo.String(c.Do("RPOP", key))
	if err != nil {
		if err == redigo.ErrNil {
			err = nil
			listVal = ""
			return
		}
	}
	listVal = reply
	return
}

//RedLock 获得一个分布式锁，针对在同一个进程加锁并且解锁的场景
func RedLock(key string, timeoutMs int) (mutex *redsync.Mutex, err error) {
	redLock := redsync.New([]redsync.Pool{pool})
	mutex = redLock.NewMutex(key,
		redsync.SetExpiry(time.Duration(timeoutMs)*time.Millisecond), //锁的过期时间为timeoutMs
		redsync.SetRetryDelay(time.Duration(10)*time.Millisecond)) //若是第一次获取不到锁，delay多久重试
	return
}

//DoTryLock 分布式锁加锁，针对在一个进程加锁，另一个进程解锁的情况
func DoTryLock(key string, timeoutMs int) (err error) {
	c := Conn()
	defer c.Close()
	val := utils.Num2Str(utils.Now())
	_, err = redigo.String(c.Do("SET", key, val, "NX", "PX", timeoutMs))
	if err != nil {
		if err == redigo.ErrNil {
			err = fmt.Errorf("key=" + key + " been locked")
		}
	}
	return
}

//DoUnLock 分布式锁解锁
func DoUnLock(key string) (err error) {
	err = RedisDel(key)
	return
}

//RedisSAdd sadd
func RedisSAdd(key string, mem ...interface{}) (err error) {
	c := Conn()
	defer c.Close()
	ms := make([]interface{}, 0)
	ms = append(ms, key)
	for _, m := range mem {
		ms = append(ms, m)
	}
	_, err = c.Do("SADD", ms...)
	return
}

//RedisSISMEMBER SISMEMBER
func RedisSISMEMBER(key, mem string) (exist bool, err error) {
	c := Conn()
	defer c.Close()
	exist, err = redigo.Bool(c.Do("SISMEMBER", key, mem))
	return
}

//RedisPub publish key&value
func RedisPub(key, val string) (err error) {
	c := Conn()
	defer c.Close()
	subNum, err := redigo.Int(c.Do("PUBLISH", key, val))
	if err != nil {
		logger.MainLogger.Error("pub key=" + key + ",val=" + val + ",err=" + err.Error())
		return
	}
	logger.MainLogger.Error("pub key=" + key + ",val=" + val + ",success. sub num==" + utils.Num2Str(subNum))
	return
}

//RedisSub key
func RedisSub(
	ctx context.Context,
	onMessage func(channel string, data []byte),
	onError func(err error, channels ...string),
	channels ...string) {

	channelList := []string{}
	for _, channel := range channels {
		logger.MainLogger.Error("RedisSub channels=" + channel)
		channelList = append(channelList, channel)

	}
	c := Conn()
	defer c.Close()
	psc := redigo.PubSubConn{Conn: c}
	err := psc.Subscribe(redigo.Args{}.AddFlat(channels)...)
	if err != nil {
		logger.MainLogger.Error(err.Error())
		duration := utils.RandIntn(5, 30)
		time.AfterFunc(time.Millisecond*time.Duration(duration), func() {
			onError(err, channels...)
		})
		return
	}
	done := make(chan error, 1)
	logger.MainLogger.Error("RedisSub loop channels")
	go func() {
		for {
			if psc.Conn.Err() != nil {
				logger.MainLogger.Error(psc.Conn.Err().Error())
			}
			switch v := psc.Receive().(type) {
			case redigo.Message:
				logger.MainLogger.Debug("Recv channel=" + v.Channel + ",msg=" + string(v.Data))
				onMessage(v.Channel, v.Data)
				logger.MainLogger.Debug("process done for channel=" + v.Channel + ",msg=" + string(v.Data))
			case redigo.Subscription:
				logger.MainLogger.Debug("RedisSub=" + v.Channel + ",kind=" + v.Kind + ",count=" + utils.Num2Str(v.Count))
				switch v.Count {
				case len(channels):
					logger.MainLogger.Debug("sub channels=" + strings.Join(channelList, "|"))
				case 0:
					logger.MainLogger.Debug("has not sub for channels=" + strings.Join(channelList, "|"))
					done <- nil
					return
				}
			case redigo.Pong:
				logger.MainLogger.Error(v.Data)
			case error:
				logger.MainLogger.Error(v.Error())
				done <- v
				return
			}
		}
	}()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	/*
		两个goroutine分别为｛主goroutine｝和｛接收goroutine｝
		两种关闭处理
		1.异常来自loop循环：其发现了ping不通或者上层主动关闭｛主goroutine｝，则
			a.结束｛主goroutine｝的循环
			b.unsub
			c.｛接收goroutine｝检测到unsub，查看是否已经全部都unsub，若是全部，则结束goroutine
			d.｛主goroutine｝收到管道的关闭信号，做清理善后操作，结束｛主goroutine｝
		2.异常来自｛接收goroutine｝
			a.｛接收goroutine｝检测到错误，则在管道中塞入错误信息，结束goroutine
			b.｛主goroutine｝收到｛接收goroutine｝的管道信息
			c.｛主goroutine｝做清理善后工作，直接返回
	*/
loop:
	for err == nil {
		select {
		case <-ticker.C:
			if err = psc.Ping("ping"); err != nil {
				logger.MainLogger.Error(err.Error())
				break loop
			}
		case <-ctx.Done():
			break loop
		case err = <-done: //此错误为｛接收goroutine｝中的case error:来的,并不是done<-nil中来的，即“异常来自｛接收goroutine｝的”处理
			logger.MainLogger.Error(err.Error())
			onError(err, channels...)
			return
		}
	}
	err1 := psc.Unsubscribe()
	if err1 != nil {
		logger.MainLogger.Error(err1.Error())
	}
	err = <-done //等待unsub的回应
	onError(err, channels...)
	logger.MainLogger.Error("done for RedisSub")
	return
}
