package redis

import (
	"encoding/json"
	"fmt"
	redigo "github.com/gomodule/redigo/redis"
	"time"
)

type Redis struct {
	name   string
	pool   *redigo.Pool
	prefix string // 前缀
	slow   int    // 慢查询时间
}

func newRedis(name string, pool *redigo.Pool, prefix string) *Redis {
	return &Redis{
		name:   name,
		pool:   pool,
		prefix: prefix,
	}
}

// Close 关闭Redis连接
func (r *Redis) Close() error {
	return r.pool.Close()
}

// Do 执行redis命令并返回结果。执行时从连接池获取连接并在执行完命令后关闭连接。
func (r *Redis) Do(commandName string, args ...interface{}) (interface{}, error) {
	conn := r.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	return conn.Do(commandName, args...)
}

// Int is a helper that converts a command reply to an integer
func (r *Redis) Int(reply interface{}, err error) (int, error) {
	return redigo.Int(reply, err)
}

// Int64 is a helper that converts a command reply to 64 bit integer
func (r *Redis) Int64(reply interface{}, err error) (int64, error) {
	return redigo.Int64(reply, err)
}

// Float64 is a helper that converts a command reply to 64 bit float
func (r *Redis) Float64(reply interface{}, err error) (float64, error) {
	return redigo.Float64(reply, err)
}

// String is a helper that converts a command reply to a string
func (r *Redis) String(reply interface{}, err error) (string, error) {
	return redigo.String(reply, err)
}

// Strings is a helper that converts an array command reply to a []string
func (r *Redis) Strings(reply interface{}, err error) ([]string, error) {
	return redigo.Strings(reply, err)
}

// Bool is a helper that converts a command reply to a boolean
func (r *Redis) Bool(reply interface{}, err error) (bool, error) {
	return redigo.Bool(reply, err)
}

// Get 获取键值。一般不直接使用该值，而是配合下面的工具类方法获取具体类型的值，或者直接使用github.com/gomodule/redigo/redis包的工具方法。
func (r *Redis) Get(key string) (interface{}, error) {
	return r.Do("GET", r.getKey(key))
}

// GetString 获取string类型的键值
func (r *Redis) GetString(key string) (string, error) {
	return r.String(r.Get(key))
}

// GetInt 获取int类型的键值
func (r *Redis) GetInt(key string) (int, error) {
	return r.Int(r.Get(key))
}

// GetInt64 获取int64类型的键值
func (r *Redis) GetInt64(key string) (int64, error) {
	return r.Int64(r.Get(key))
}

// GetBool 获取bool类型的键值
func (r *Redis) GetBool(key string) (bool, error) {
	return r.Bool(r.Get(key))
}

// GetObject 获取非基本类型struct的键值。在实现上，使用json的Marshal和Unmarshal做序列化存取。
func (r *Redis) GetObject(key string, val interface{}) error {
	reply, err := r.Get(key)
	return r.decode(reply, err, val)
}

// Set 存并设置有效时长。时长的单位为秒。
// 基础类型直接保存，其他用json.Marshal后转成string保存。
func (r *Redis) Set(key string, val interface{}, expire int64) error {
	value, err := r.encode(val)
	if err != nil {
		return err
	}
	if expire > 0 {
		_, err := r.Do("SETEX", r.getKey(key), expire, value)
		return err
	}
	_, err = r.Do("SET", r.getKey(key), value)
	return err
}

// SetExNx 不存在则设置有效时长。时长的单位为秒。
// 基础类型直接保存，其他用json.Marshal后转成string保存。
func (r *Redis) SetExNx(key string, val interface{}, expire int64) error {
	value, err := r.encode(val)
	if err != nil {
		return err
	}
	_, err = r.Do("SET", key, value, "EX", expire, "NX")
	return err
}

// Exists 检查键是否存在
func (r *Redis) Exists(key string) (bool, error) {
	return r.Bool(r.Do("EXISTS", r.getKey(key)))
}

// Del 删除键
func (r *Redis) Del(key string) error {
	_, err := r.Do("DEL", r.getKey(key))
	return err
}

// Flush 清空当前数据库中的所有 key，慎用！
func (r *Redis) Flush() error {
	_, err := r.Do("FLUSHDB")
	return err
}

// TTL 以秒为单位。当 key 不存在时，返回 -2 。 当 key 存在但没有设置剩余生存时间时，返回 -1
func (r *Redis) TTL(key string) (ttl int64, err error) {
	return r.Int64(r.Do("TTL", r.getKey(key)))
}

// Expire 设置键过期时间，expire的单位为秒
func (r *Redis) Expire(key string, expire int64) error {
	_, err := r.Bool(r.Do("EXPIRE", r.getKey(key), expire))
	return err
}

// Incr 将 key 中储存的数字值增一
func (r *Redis) Incr(key string) (val int64, err error) {
	return r.Int64(r.Do("INCR", r.getKey(key)))
}

// IncrBy 将 key 所储存的值加上给定的增量值（increment）。
func (r *Redis) IncrBy(key string, amount int64) (val int64, err error) {
	return r.Int64(r.Do("INCRBY", r.getKey(key), amount))
}

// Decr 将 key 中储存的数字值减一。
func (r *Redis) Decr(key string) (val int64, err error) {
	return r.Int64(r.Do("DECR", r.getKey(key)))
}

// DecrBy key 所储存的值减去给定的减量值（decrement）。
func (r *Redis) DecrBy(key string, amount int64) (val int64, err error) {
	return r.Int64(r.Do("DECRBY", r.getKey(key), amount))
}

// HMSet 将一个map存到Redis hash，同时设置有效期，单位：秒
// Example:
//
// ```golang
// m := make(map[string]interface{})
// m["name"] = "corel"
// m["age"] = 23
// err := r.HMSet("user", m, 10)
// ```
func (r *Redis) HMSet(key string, val interface{}, expire int) error {
	conn := r.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	err := conn.Send("HMSET", redigo.Args{}.Add(r.getKey(key)).AddFlat(val)...)
	if err != nil {
		return err
	}
	if expire > 0 {
		err = conn.Send("EXPIRE", r.getKey(key), int64(expire))
	}
	if err != nil {
		return err
	}
	conn.Flush()
	_, err = conn.Receive()
	return err
}

/** Redis hash 是一个string类型的field和value的映射表，hash特别适合用于存储对象。 **/

// HSet 将哈希表 key 中的字段 field 的值设为 val
// Example:
//
// ```golang
// _, err := r.HSet("user", "age", 23)
// ```
func (r *Redis) HSet(key, field string, val interface{}) (interface{}, error) {
	value, err := r.encode(val)
	if err != nil {
		return nil, err
	}
	return r.Do("HSET", r.getKey(key), field, value)
}

// HGet 获取存储在哈希表中指定字段的值
// Example:
//
// ```golang
// val, err := r.HGet("user", "age")
// ```
func (r *Redis) HGet(key, field string) (reply interface{}, err error) {
	reply, err = r.Do("HGET", r.getKey(key), field)
	return
}

// HGetString HGet的工具方法，当字段值为字符串类型时使用
func (r *Redis) HGetString(key, field string) (reply string, err error) {
	reply, err = r.String(r.HGet(key, field))
	return
}

// HGetInt HGet的工具方法，当字段值为int类型时使用
func (r *Redis) HGetInt(key, field string) (reply int, err error) {
	reply, err = r.Int(r.HGet(key, field))
	return
}

// HGetInt64 HGet的工具方法，当字段值为int64类型时使用
func (r *Redis) HGetInt64(key, field string) (reply int64, err error) {
	reply, err = r.Int64(r.HGet(key, field))
	return
}

// HGetBool HGet的工具方法，当字段值为bool类型时使用
func (r *Redis) HGetBool(key, field string) (reply bool, err error) {
	reply, err = r.Bool(r.HGet(key, field))
	return
}

// HGetObject HGet的工具方法，当字段值为非基本类型的stuct时使用
func (r *Redis) HGetObject(key, field string, val interface{}) error {
	reply, err := r.HGet(key, field)
	return r.decode(reply, err, val)
}

// HGetAll HGetAll("key", &val)
func (r *Redis) HGetAll(key string, val interface{}) error {
	v, err := redigo.Values(r.Do("HGETALL", r.getKey(key)))
	if err != nil {
		return err
	}

	if err := redigo.ScanStruct(v, val); err != nil {
		fmt.Println(err)
	}
	//fmt.Printf("%+v\n", val)
	return err
}

/**
Redis列表是简单的字符串列表，按照插入顺序排序。你可以添加一个元素到列表的头部（左边）或者尾部（右边）
**/

// BLPop 它是 LPOP 命令的阻塞版本，当给定列表内没有任何元素可供弹出的时候，连接将被 BLPOP 命令阻塞，直到等待超时或发现可弹出元素为止。
// 超时参数 timeout 接受一个以秒为单位的数字作为值。超时参数设为 0 表示阻塞时间可以无限期延长(block indefinitely) 。
func (r *Redis) BLPop(key string, timeout int) (interface{}, error) {
	values, err := redigo.Values(r.Do("BLPOP", r.getKey(key), timeout))
	if err != nil {
		return nil, err
	}
	if len(values) != 2 {
		return nil, fmt.Errorf("redisgo: unexpected number of values, got %d", len(values))
	}
	return values[1], err
}

// BLPopInt BLPop的工具方法，元素类型为int时
func (r *Redis) BLPopInt(key string, timeout int) (int, error) {
	return r.Int(r.BLPop(key, timeout))
}

// BLPopInt64 BLPop的工具方法，元素类型为int64时
func (r *Redis) BLPopInt64(key string, timeout int) (int64, error) {
	return r.Int64(r.BLPop(key, timeout))
}

// BLPopString BLPop的工具方法，元素类型为string时
func (r *Redis) BLPopString(key string, timeout int) (string, error) {
	return r.String(r.BLPop(key, timeout))
}

// BLPopBool BLPop的工具方法，元素类型为bool时
func (r *Redis) BLPopBool(key string, timeout int) (bool, error) {
	return r.Bool(r.BLPop(key, timeout))
}

// BLPopObject BLPop的工具方法，元素类型为object时
func (r *Redis) BLPopObject(key string, timeout int, val interface{}) error {
	reply, err := r.BLPop(key, timeout)
	return r.decode(reply, err, val)
}

// BRPop 它是 RPOP 命令的阻塞版本，当给定列表内没有任何元素可供弹出的时候，连接将被 BRPOP 命令阻塞，直到等待超时或发现可弹出元素为止。
// 超时参数 timeout 接受一个以秒为单位的数字作为值。超时参数设为 0 表示阻塞时间可以无限期延长(block indefinitely) 。
func (r *Redis) BRPop(key string, timeout int) (interface{}, error) {
	values, err := redigo.Values(r.Do("BRPOP", r.getKey(key), timeout))
	if err != nil {
		return nil, err
	}
	if len(values) != 2 {
		return nil, fmt.Errorf("redisgo: unexpected number of values, got %d", len(values))
	}
	return values[1], err
}

// BRPopInt BRPop的工具方法，元素类型为int时
func (r *Redis) BRPopInt(key string, timeout int) (int, error) {
	return r.Int(r.BRPop(key, timeout))
}

// BRPopInt64 BRPop的工具方法，元素类型为int64时
func (r *Redis) BRPopInt64(key string, timeout int) (int64, error) {
	return r.Int64(r.BRPop(key, timeout))
}

// BRPopString BRPop的工具方法，元素类型为string时
func (r *Redis) BRPopString(key string, timeout int) (string, error) {
	return r.String(r.BRPop(key, timeout))
}

// BRPopBool BRPop的工具方法，元素类型为bool时
func (r *Redis) BRPopBool(key string, timeout int) (bool, error) {
	return r.Bool(r.BRPop(key, timeout))
}

// BRPopObject BRPop的工具方法，元素类型为object时
func (r *Redis) BRPopObject(key string, timeout int, val interface{}) error {
	reply, err := r.BRPop(key, timeout)
	return r.decode(reply, err, val)
}

// LPop 移出并获取列表中的第一个元素（表头，左边）
func (r *Redis) LPop(key string) (interface{}, error) {
	return r.Do("LPOP", r.getKey(key))
}

// LPopInt 移出并获取列表中的第一个元素（表头，左边），元素类型为int
func (r *Redis) LPopInt(key string) (int, error) {
	return r.Int(r.LPop(key))
}

// LPopInt64 移出并获取列表中的第一个元素（表头，左边），元素类型为int64
func (r *Redis) LPopInt64(key string) (int64, error) {
	return r.Int64(r.LPop(key))
}

// LPopString 移出并获取列表中的第一个元素（表头，左边），元素类型为string
func (r *Redis) LPopString(key string) (string, error) {
	return r.String(r.LPop(key))
}

// LPopBool 移出并获取列表中的第一个元素（表头，左边），元素类型为bool
func (r *Redis) LPopBool(key string) (bool, error) {
	return r.Bool(r.LPop(key))
}

// LPopObject 移出并获取列表中的第一个元素（表头，左边），元素类型为非基本类型的struct
func (r *Redis) LPopObject(key string, val interface{}) error {
	reply, err := r.LPop(key)
	return r.decode(reply, err, val)
}

// RPop 移出并获取列表中的最后一个元素（表尾，右边）
func (r *Redis) RPop(key string) (interface{}, error) {
	return r.Do("RPOP", r.getKey(key))
}

// RPopInt 移出并获取列表中的最后一个元素（表尾，右边），元素类型为int
func (r *Redis) RPopInt(key string) (int, error) {
	return r.Int(r.RPop(key))
}

// RPopInt64 移出并获取列表中的最后一个元素（表尾，右边），元素类型为int64
func (r *Redis) RPopInt64(key string) (int64, error) {
	return r.Int64(r.RPop(key))
}

// RPopString 移出并获取列表中的最后一个元素（表尾，右边），元素类型为string
func (r *Redis) RPopString(key string) (string, error) {
	return r.String(r.RPop(key))
}

// RPopBool 移出并获取列表中的最后一个元素（表尾，右边），元素类型为bool
func (r *Redis) RPopBool(key string) (bool, error) {
	return r.Bool(r.RPop(key))
}

// RPopObject 移出并获取列表中的最后一个元素（表尾，右边），元素类型为非基本类型的struct
func (r *Redis) RPopObject(key string, val interface{}) error {
	reply, err := r.RPop(key)
	return r.decode(reply, err, val)
}

// LPush 将一个值插入到列表头部
func (r *Redis) LPush(key string, member interface{}) error {
	value, err := r.encode(member)
	if err != nil {
		return err
	}
	_, err = r.Do("LPUSH", r.getKey(key), value)
	return err
}

// RPush 将一个值插入到列表尾部
func (r *Redis) RPush(key string, member interface{}) error {
	value, err := r.encode(member)
	if err != nil {
		return err
	}
	_, err = r.Do("RPUSH", r.getKey(key), value)
	return err
}

// LREM 根据参数 count 的值，移除列表中与参数 member 相等的元素。
// count 的值可以是以下几种：
// count > 0 : 从表头开始向表尾搜索，移除与 member 相等的元素，数量为 count 。
// count < 0 : 从表尾开始向表头搜索，移除与 member 相等的元素，数量为 count 的绝对值。
// count = 0 : 移除表中所有与 member 相等的值。
// 返回值：被移除元素的数量。
func (r *Redis) LREM(key string, count int, member interface{}) (int, error) {
	return r.Int(r.Do("LREM", r.getKey(key), count, member))
}

// LLen 获取列表的长度
func (r *Redis) LLen(key string) (int64, error) {
	return r.Int64(r.Do("RPOP", r.getKey(key)))
}

// LRange 返回列表 key 中指定区间内的元素，区间以偏移量 start 和 stop 指定。
// 下标(index)参数 start 和 stop 都以 0 为底，也就是说，以 0 表示列表的第一个元素，以 1 表示列表的第二个元素，以此类推。
// 你也可以使用负数下标，以 -1 表示列表的最后一个元素， -2 表示列表的倒数第二个元素，以此类推。
// 和编程语言区间函数的区别：end 下标也在 LRANGE 命令的取值范围之内(闭区间)。
func (r *Redis) LRange(key string, start, end int) (interface{}, error) {
	return r.Do("LRANGE", r.getKey(key), start, end)
}

/**
Redis 有序集合和集合一样也是string类型元素的集合,且不允许重复的成员。
不同的是每个元素都会关联一个double类型的分数。redis正是通过分数来为集合中的成员进行从小到大的排序。
有序集合的成员是唯一的,但分数(score)却可以重复。
集合是通过哈希表实现的，所以添加，删除，查找的复杂度都是O(1)。
**/

// ZAdd 将一个 member 元素及其 score 值加入到有序集 key 当中。
func (r *Redis) ZAdd(key string, score int64, member string) (reply interface{}, err error) {
	return r.Do("ZADD", r.getKey(key), score, member)
}

// ZRem 移除有序集 key 中的一个成员，不存在的成员将被忽略。
func (r *Redis) ZRem(key string, member string) (reply interface{}, err error) {
	return r.Do("ZREM", r.getKey(key), member)
}

// ZScore 返回有序集 key 中，成员 member 的 score 值。 如果 member 元素不是有序集 key 的成员，或 key 不存在，返回 nil 。
func (r *Redis) ZScore(key string, member string) (int64, error) {
	return r.Int64(r.Do("ZSCORE", r.getKey(key), member))
}

// ZRank 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从小到大)顺序排列。score 值最小的成员排名为 0
func (r *Redis) ZRank(key, member string) (int64, error) {
	return r.Int64(r.Do("ZRANK", r.getKey(key), member))
}

// ZRevrank 返回有序集中成员的排名。其中有序集成员按分数值递减(从大到小)排序。分数值最大的成员排名为 0 。
func (r *Redis) ZRevrank(key, member string) (int64, error) {
	return r.Int64(r.Do("ZREVRANK", r.getKey(key), member))
}

// ZRange 返回有序集中，指定区间内的成员。其中成员的位置按分数值递增(从小到大)来排序。具有相同分数值的成员按字典序(lexicographical order )来排列。
// 以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，以此类推。或 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
func (r *Redis) ZRange(key string, from, to int64) (map[string]int64, error) {
	return redigo.Int64Map(r.Do("ZRANGE", r.getKey(key), from, to, "WITHSCORES"))
}

// ZRevrange 返回有序集中，指定区间内的成员。其中成员的位置按分数值递减(从大到小)来排列。具有相同分数值的成员按字典序(lexicographical order )来排列。
// 以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，以此类推。或 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
func (r *Redis) ZRevrange(key string, from, to int64) (map[string]int64, error) {
	return redigo.Int64Map(r.Do("ZREVRANGE", r.getKey(key), from, to, "WITHSCORES"))
}

// ZRangeByScore 返回有序集合中指定分数区间的成员列表。有序集成员按分数值递增(从小到大)次序排列。
// 具有相同分数值的成员按字典序来排列
func (r *Redis) ZRangeByScore(key string, from, to, offset int64, count int) (map[string]int64, error) {
	return redigo.Int64Map(r.Do("ZRANGEBYSCORE", r.getKey(key), from, to, "WITHSCORES", "LIMIT", offset, count))
}

// ZRevrangeByScore 返回有序集中指定分数区间内的所有的成员。有序集成员按分数值递减(从大到小)的次序排列。
// 具有相同分数值的成员按字典序来排列
func (r *Redis) ZRevrangeByScore(key string, from, to, offset int64, count int) (map[string]int64, error) {
	return redigo.Int64Map(r.Do("ZREVRANGEBYSCORE", r.getKey(key), from, to, "WITHSCORES", "LIMIT", offset, count))
}

/**
Redis 发布订阅(pub/sub)是一种消息通信模式：发送者(pub)发送消息，订阅者(sub)接收消息。
Redis 客户端可以订阅任意数量的频道。
当有新消息通过 PUBLISH 命令发送给频道 channel 时， 这个消息就会被发送给订阅它的所有客户端。
**/

// Publish 将信息发送到指定的频道，返回接收到信息的订阅者数量
func (r *Redis) Publish(channel, message string) (int, error) {
	return r.Int(r.Do("PUBLISH", channel, message))
}

// Subscribe 订阅给定的一个或多个频道的信息。
// 支持redis服务停止或网络异常等情况时，自动重新订阅。
// 一般的程序都是启动后开启一些固定channel的订阅，也不会动态的取消订阅，这种场景下可以使用本方法。
// 复杂场景的使用可以直接参考 https://godoc.org/github.com/gomodule/redigo/redis#hdr-Publish_and_Subscribe
func (r *Redis) Subscribe(onMessage func(channel string, data []byte) error, channels ...string) error {
	conn := r.pool.Get()
	psc := redigo.PubSubConn{Conn: conn}
	err := psc.Subscribe(redigo.Args{}.AddFlat(channels)...)
	// 如果订阅失败，休息1秒后重新订阅（比如当redis服务停止服务或网络异常）
	if err != nil {
		fmt.Println(err)
		time.Sleep(time.Second)
		return r.Subscribe(onMessage, channels...)
	}
	quit := make(chan int, 1)

	// 处理消息
	go func() {
		for {
			switch v := psc.Receive().(type) {
			case redigo.Message:
				// fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
				go onMessage(v.Channel, v.Data)
			case redigo.Subscription:
				fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
			case error:
				quit <- 1
				fmt.Println(err)
				return
			}
		}
	}()

	// 异常情况下自动重新订阅
	go func() {
		<-quit
		time.Sleep(time.Second)
		psc.Close()
		r.Subscribe(onMessage, channels...)
	}()
	return err
}

/**
GEO 地理位置
*/

// GeoOptions 用于GEORADIUS和GEORADIUSBYMEMBER命令的参数
type GeoOptions struct {
	WithCoord bool
	WithDist  bool
	WithHash  bool
	Order     string // ASC从近到远，DESC从远到近
	Count     int
}

// GeoResult 用于GEORADIUS和GEORADIUSBYMEMBER命令的查询结果
type GeoResult struct {
	Name      string
	Longitude float64
	Latitude  float64
	Dist      float64
	Hash      int64
}

// GeoAdd 将给定的空间元素（纬度、经度、名字）添加到指定的键里面，这些数据会以有序集合的形式被储存在键里面，所以删除可以使用`ZREM`。
func (r *Redis) GeoAdd(key string, longitude, latitude float64, member string) error {
	_, err := r.Int(r.Do("GEOADD", r.getKey(key), longitude, latitude, member))
	return err
}

// GeoPos 从键里面返回所有给定位置元素的位置（经度和纬度）。
func (r *Redis) GeoPos(key string, members ...interface{}) ([]*[2]float64, error) {
	args := redigo.Args{}
	args = args.Add(r.getKey(key))
	args = args.Add(members...)
	return redigo.Positions(r.Do("GEOPOS", args...))
}

// GeoDist 返回两个给定位置之间的距离。
// 如果两个位置之间的其中一个不存在， 那么命令返回空值。
// 指定单位的参数 unit 必须是以下单位的其中一个：
// m 表示单位为米。
// km 表示单位为千米。
// mi 表示单位为英里。
// ft 表示单位为英尺。
// 如果用户没有显式地指定单位参数， 那么 GEODIST 默认使用米作为单位。
func (r *Redis) GeoDist(key string, member1, member2, unit string) (float64, error) {
	_, err := r.Float64(r.Do("GEODIST", r.getKey(key), member1, member2, unit))
	return 0, err
}

// GeoRadius 以给定的经纬度为中心， 返回键包含的位置元素当中， 与中心的距离不超过给定最大距离的所有位置元素。
func (r *Redis) GeoRadius(key string, longitude, latitude, radius float64, unit string, options GeoOptions) ([]*GeoResult, error) {
	args := redigo.Args{}
	args = args.Add(r.getKey(key), longitude, latitude, radius, unit)
	if options.WithDist {
		args = args.Add("WITHDIST")
	}
	if options.WithCoord {
		args = args.Add("WITHCOORD")
	}
	if options.WithHash {
		args = args.Add("WITHHASH")
	}
	if options.Order != "" {
		args = args.Add(options.Order)
	}
	if options.Count > 0 {
		args = args.Add("Count", options.Count)
	}

	reply, err := r.Do("GEORADIUS", args...)
	return r.toGeoResult(reply, err, options)
}

// GeoRadiusByMember 这个命令和 GEORADIUS 命令一样， 都可以找出位于指定范围内的元素， 但是 GEORADIUSBYMEMBER 的中心点是由给定的位置元素决定的， 而不是像 GEORADIUS 那样， 使用输入的经度和纬度来决定中心点。
func (r *Redis) GeoRadiusByMember(key string, member string, radius float64, unit string, options GeoOptions) ([]*GeoResult, error) {
	args := redigo.Args{}
	args = args.Add(r.getKey(key), member, radius, unit)
	if options.WithDist {
		args = args.Add("WITHDIST")
	}
	if options.WithCoord {
		args = args.Add("WITHCOORD")
	}
	if options.WithHash {
		args = args.Add("WITHHASH")
	}
	if options.Order != "" {
		args = args.Add(options.Order)
	}
	if options.Count > 0 {
		args = args.Add("Count", options.Count)
	}

	reply, err := r.Do("GEORADIUSBYMEMBER", args...)
	return r.toGeoResult(reply, err, options)
}

// GeoHash 返回一个或多个位置元素的 Geohash 表示。
func (r *Redis) GeoHash(key string, members ...interface{}) ([]string, error) {
	args := redigo.Args{}
	args = args.Add(r.getKey(key))
	args = args.Add(members...)
	return r.Strings(r.Do("GEOHASH", args...))
}

func (r *Redis) toGeoResult(reply interface{}, err error, options GeoOptions) ([]*GeoResult, error) {
	values, err := redigo.Values(reply, err)
	if err != nil {
		return nil, err
	}
	results := make([]*GeoResult, len(values))
	for i := range values {
		if values[i] == nil {
			continue
		}
		p, ok := values[i].([]interface{})
		if !ok {
			return nil, fmt.Errorf("redisgo: unexpected element type for interface slice, got type %T", values[i])
		}
		geoResult := &GeoResult{}
		pos := 0
		name, err := r.String(p[pos], nil)
		if err != nil {
			return nil, err
		}
		geoResult.Name = name
		if options.WithDist {
			pos = pos + 1
			dist, err := r.Float64(p[pos], nil)
			if err != nil {
				return nil, err
			}
			geoResult.Dist = dist
		}
		if options.WithHash {
			pos = pos + 1
			hash, err := r.Int64(p[pos], nil)
			if err != nil {
				return nil, err
			}
			geoResult.Hash = hash
		}
		if options.WithCoord {
			pos = pos + 1
			pp, ok := p[pos].([]interface{})
			if !ok {
				return nil, fmt.Errorf("redisgo: unexpected element type for interface slice, got type %T", p[i])
			}
			if len(pp) > 0 {
				lat, err := r.Float64(pp[0], nil)
				if err != nil {
					return nil, err
				}
				lon, err := r.Float64(pp[1], nil)
				if err != nil {
					return nil, err
				}
				geoResult.Latitude = lat
				geoResult.Longitude = lon
			}
		}
		results[i] = geoResult
	}
	return results, nil
}

// getKey 将健名加上指定的前缀。
func (r *Redis) getKey(key string) string {
	return r.prefix + key
}

// encode 序列化要保存的值
func (r *Redis) encode(val interface{}) (interface{}, error) {
	var value interface{}
	switch v := val.(type) {
	case string, int, uint, int8, int16, int32, int64, float32, float64, bool:
		value = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		value = string(b)
	}
	return value, nil
}

// decode 反序列化保存的struct对象
func (r *Redis) decode(reply interface{}, err error, val interface{}) error {
	str, err := r.String(reply, err)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(str), val)
}
