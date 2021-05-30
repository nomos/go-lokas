package redisclient

import (
	"github.com/gomodule/redigo/redis"
	"github.com/nomos/go-log/log"
	"reflect"
	"time"
)

type Client struct {
	*redis.Pool
}

func NewClient(addr string) (*Client,error) {
	ret := &Client{
		Pool : &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
			Dial: func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
		},
	}
	_, err := ret.ping()
	return ret, err
}

func (this *Client) ping() (bool, error) {
	conn := this.Pool.Get()
	defer conn.Close()
	data, err := conn.Do("PING")
	if err != nil || data == nil {
		return false, err
	}
	return (data == "PONG"), nil
}

func (this *Client) Execute(cmd string, args ...interface{}) *RedisReply {
	conn := this.Pool.Get()
	defer conn.Close()
	res, err := conn.Do(cmd, args...)
	log.Infof(cmd, args, res, reflect.TypeOf(res))
	if err != nil {
		log.Error(err.Error())
	}
	return NewReply(res, err)
}

//String
//
func (this *Client) Exists(k string) *RedisReply {
	return this.Execute("EXISTS", k)
}

//设置指定 key 的值
func (this *Client) Set(k string, v interface{}) *RedisReply {
	return this.Execute("SET", k, v)
}

//获取指定 key 的值。
func (this *Client) Get(k string) *RedisReply {
	return this.Execute("GET", k)
}

//返回 key 中字符串值的子字符
func (this *Client) GetRange(k string, start, end int) *RedisReply {
	return this.Execute("GETRANGE", k, start, end)
}

//将给定 key 的值设为 value ，并返回 key 的旧值(old value)。
func (this *Client) GetSet(k string, v interface{}) *RedisReply {
	return this.Execute("GETSET", k, v)
}

//对 key 所储存的字符串值，获取指定偏移量上的位(bit)。
func (this *Client) GetBit(k string, offset int) *RedisReply {
	return this.Execute("GETSET", k, offset)
}

//获取所有(一个或多个)给定 key 的值。
func (this *Client) MGet(keys ...interface{}) *RedisReply {
	return this.Execute("MGET", keys...)
}

//
func (this *Client) Del(k ...interface{}) *RedisReply {
	return this.Execute("DEL", k...)
}

//对 key 所储存的字符串值，设置或清除指定偏移量上的位(bit)。
func (this *Client) SetBit(k string, offset int, v int) *RedisReply {
	return this.Execute("SETBIT", k, offset, v)
}

//将值 value 关联到 key ，并将 key 的过期时间设为 seconds (以秒为单位)。
func (this *Client) SetEx(k string, seconds float64) *RedisReply {
	return this.Execute("SETEX", k, seconds)
}

//只有在 key 不存在时设置 key 的值。
func (this *Client) SetNx(k string, v interface{}) *RedisReply {
	return this.Execute("SETNX", k, v)
}

//用 value 参数覆写给定 key 所储存的字符串值，从偏移量 offset 开始。
func (this *Client) SetRange(k string, offset int, v interface{}) *RedisReply {
	return this.Execute("SETRANGE", k, offset, v)
}

//返回 key 所储存的字符串值的长度。
func (this *Client) StrLen(k string) *RedisReply {
	return this.Execute("STRLEN", k)
}

//同时设置一个或多个 key-value 对。
func (this *Client) MSet(args ...interface{}) *RedisReply {
	return this.Execute("MSET", args...)
}

//同时设置一个或多个 key-value 对，当且仅当所有给定 key 都不存在。
func (this *Client) MSetNx(args ...interface{}) *RedisReply {
	return this.Execute("MSETNX", args...)
}

//这个命令和 SETEX 命令相似，但它以毫秒为单位设置 key 的生存时间，而不是像 SETEX 命令那样，以秒为单位。
func (this *Client) PSetEx(k string, milliseconds float64) *RedisReply {
	return this.Execute("PSETEX", k, milliseconds)
}

//将 key 中储存的数字值增一。
func (this *Client) Incr(k string) *RedisReply {
	return this.Execute("INCR", k)
}

//将 key 所储存的值加上给定的增量值（increment） 。
func (this *Client) IncrBy(k string, incr int) *RedisReply {
	return this.Execute("INCRBY", k, incr)
}

//将 key 所储存的值加上给定的浮点增量值（increment） 。
func (this *Client) IncrByFloat(k string, incr float64) *RedisReply {
	return this.Execute("INCRBYFLOAT", k, incr)
}

//将 key 中储存的数字值减一。
func (this *Client) Decr(k string) *RedisReply {
	return this.Execute("DECR", k)
}

//key 所储存的值减去给定的减量值（decrement） 。
func (this *Client) DecrBy(k string, decr int) *RedisReply {
	return this.Execute("DECRBY", k, decr)
}

//key 所储存的值减去给定的浮点减量值（decrement） 。
func (this *Client) DecrByFloat(k string, decr float64) *RedisReply {
	return this.Execute("DECRBYFLOAT", k, decr)
}

//如果 key 已经存在并且是一个字符串， APPEND 命令将指定的 value 追加到该 key 原来值（value）的末尾。
func (this *Client) Append(k string, v string) *RedisReply {
	return this.Execute("APPEND", k, v)
}

//Hash
//删除一个或多个哈希表字段
func (this *Client) HDel(k string, f ...interface{}) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	args = append(args, f...)
	return this.Execute("HDEL", args)
}

//查看哈希表 key 中，指定的字段是否存在。
func (this *Client) HExists(k string, f string) *RedisReply {
	return this.Execute("HEXISTS", k, f)
}

//获取存储在哈希表中指定字段的值。
func (this *Client) HGet(k string, f string) *RedisReply {
	return this.Execute("HGET", k, f)
}

//获取在哈希表中指定 key 的所有字段和值
func (this *Client) HGetAll(k string) *RedisReply {
	return this.Execute("HGETALL", k)
}

//为哈希表 key 中的指定字段的整数值加上增量 increment 。
func (this *Client) HIncrBy(k string, f string, incre int) *RedisReply {
	return this.Execute("HINCRBY", k, f, incre)
}

//为哈希表 key 中的指定字段的浮点数值加上增量 increment 。
func (this *Client) HIncrByFloat(k string, f string, incre float64) *RedisReply {
	return this.Execute("HINCRBYFLOAT", k, f, incre)
}

//获取所有哈希表中的字段
func (this *Client) HKeys(k string) *RedisReply {
	return this.Execute("HKEYS", k)
}

//获取哈希表中字段的数量
func (this *Client) HLen(k string) *RedisReply {
	return this.Execute("HLEN", k)
}

//获取所有给定字段的值
func (this *Client) HMGet(args ...interface{}) *RedisReply {
	return this.Execute("HMGET", args...)
}

//同时将多个 field-value (域-值)对设置到哈希表 key 中。
func (this *Client) HMSet(k string, args ...interface{}) *RedisReply {
	args2 := []interface{}{}
	args2 = append(args2, k)
	args2 = append(args2, args...)
	return this.Execute("HMSET", args2...)
}

//将哈希表 key 中的字段 field 的值设为 value 。
func (this *Client) HSet(k string, f string, v interface{}) *RedisReply {
	return this.Execute("HSET", k, f, v)
}

//只有在字段 field 不存在时，设置哈希表字段的值。
func (this *Client) HSetNx(k string, f string, v interface{}) *RedisReply {
	return this.Execute("HSETNX", k, f, v)
}

//获取哈希表中所有值。
func (this *Client) HVals(k string) *RedisReply {
	return this.Execute("HVALS", k)
}

//List
//移出并获取列表的第一个元素， 如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止。
func (this *Client) BLPop(args ...interface{}) *RedisReply {
	return this.Execute("BLPOP", args...)
}

//移出并获取列表的最后一个元素， 如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止。
func (this *Client) BRPop(args ...interface{}) *RedisReply {
	return this.Execute("BRPOP", args...)
}

//从列表中弹出一个值，将弹出的元素插入到另外一个列表中并返回它； 如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止。
func (this *Client) BRPopLPush(source, dest string, args ...interface{}) *RedisReply {
	args2 := []interface{}{}
	args2 = append(args2, source, dest)
	args2 = append(args2, args...)
	return this.Execute("BRPOPLPUSH", args2...)
}

//通过索引获取列表中的元素
func (this *Client) LIndex(k string, index int) *RedisReply {
	return this.Execute("LINDEX", k, index)
}

//在列表的元素前或者后插入元素
func (this *Client) LInsert(k string, before bool, pivot int, v interface{}) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	if before {
		args = append(args, "BEFORE")
	} else {
		args = append(args, "AFTER")
	}
	args = append(args, pivot)
	args = append(args, v)
	return this.Execute("LINSERT", args...)
}

//获取列表长度
func (this *Client) LLen(k string) *RedisReply {
	return this.Execute("LLEN", k)
}

//移出并获取列表的第一个元素
func (this *Client) LPop(k string) *RedisReply {
	return this.Execute("LPOP", k)
}

//将一个或多个值插入到列表头部
func (this *Client) LPush(k string, v ...interface{}) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	args = append(args, v...)
	return this.Execute("LPUSH", args...)
}

//将一个值插入到已存在的列表头部
func (this *Client) LPushX(k string, v interface{}) *RedisReply {
	return this.Execute("LPUSHX", k, v)
}

//获取列表指定范围内的元素
func (this *Client) LRange(k string, start, end int) *RedisReply {
	return this.Execute("LRANGE", k, start, end)
}

//移除列表元素
func (this *Client) LRem(k string, count int, v interface{}) *RedisReply {
	return this.Execute("LREM", k, count, v)
}

//通过索引设置列表元素的值
func (this *Client) LSet(k string, index int, v interface{}) *RedisReply {
	return this.Execute("LSET", k, index, v)
}

//对一个列表进行修剪(trim)，就是说，让列表只保留指定区间内的元素，不在指定区间之内的元素都将被删除。
func (this *Client) LTrim(k string, start, end int) *RedisReply {
	return this.Execute("LTRIM", k, start, end)
}

//移除列表的最后一个元素，返回值为移除的元素。
func (this *Client) RPop(k string) *RedisReply {
	return this.Execute("RPOP", k)
}

//移除列表的最后一个元素，并将该元素添加到另一个列表并返回
func (this *Client) RPopLPush(source, dest string) *RedisReply {
	return this.Execute("RPOPLPUSH", source, dest)
}

//在列表中添加一个或多个值
func (this *Client) RPush(k string, v ...interface{}) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	args = append(args, v...)
	return this.Execute("RPUSH", args...)
}

//为已存在的列表添加值
func (this *Client) RPushX(k string, v interface{}) *RedisReply {
	return this.Execute("RPUSHX", k, v)
}

//Set
//向集合添加一个或多个成员
func (this *Client) SAdd(k string, m ...string) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	for _, v := range m {
		args = append(args, v)
	}
	return this.Execute("SADD", args...)
}

//获取集合的成员数
func (this *Client) SCard(k string, ) *RedisReply {
	return this.Execute("SCARD", k)
}

//返回第一个集合与其他集合之间的差异。
func (this *Client) SDiff(s ...string) *RedisReply {
	args := []interface{}{}
	for _, v := range s {
		args = append(args, v)
	}
	return this.Execute("SDIFF", args...)
}

//返回给定所有集合的差集并存储在 destination 中
func (this *Client) SDiffStore(dest string, s ...string) *RedisReply {
	args := []interface{}{}
	args = append(args, dest)
	for _, v := range s {
		args = append(args, v)
	}
	return this.Execute("SDIFFSTORE", args...)
}

//返回给定所有集合的交集
func (this *Client) SInter(s ...string) *RedisReply {
	args := []interface{}{}
	for _, v := range s {
		args = append(args, v)
	}
	return this.Execute("SINTER", args...)
}

//返回给定所有集合的交集并存储在 destination 中
func (this *Client) SInterStore(dest string, s ...string) *RedisReply {
	args := []interface{}{}
	args = append(args, dest)
	for _, v := range s {
		args = append(args, v)
	}
	return this.Execute("SINTERSTORE", args...)
}

//判断 member 元素是否是集合 key 的成员
func (this *Client) SIsMember(k string, m string) *RedisReply {
	return this.Execute("SISMEMBER", k, m)
}

//返回集合中的所有成员
func (this *Client) SMembers(k string) *RedisReply {
	return this.Execute("SMEMBERS", k)
}

//将 member 元素从 source 集合移动到 destination 集合
func (this *Client) SMove(source string, dest string, m string) *RedisReply {
	return this.Execute("SMOVE", source, dest, m)
}

//移除并返回集合中的一个随机元素
func (this *Client) SPop(k string, ) *RedisReply {
	return this.Execute("SPOP", k)
}

//返回集合中一个或多个随机数
func (this *Client) SRandMember(k string, count ...int) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	for _, v := range count {
		args = append(args, v)
		break
	}
	return this.Execute("SRANDMEMBER", args...)
}

//移除集合中一个或多个成员
func (this *Client) SRem(k string, m ...string) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	for _, v := range m {
		args = append(args, v)
	}
	return this.Execute("SREM", args...)
}

//返回所有给定集合的并集
func (this *Client) SUnion(s ...string) *RedisReply {
	args := []interface{}{}
	for _, v := range s {
		args = append(args, v)
	}
	return this.Execute("SUNION", args...)
}

//所有给定集合的并集存储在 destination 集合中
func (this *Client) SUnionStore(dest string, s ...string) *RedisReply {
	args := []interface{}{}
	args = append(args, dest)
	for _, v := range s {
		args = append(args, v)
	}
	return this.Execute("SUNIONSTORE", args...)
}

// sorted set

//向有序集合添加一个或多个成员，或者更新已存在成员的分数
func (this *Client) ZAdd(k string, args ...interface{}) *RedisReply {
	args2 := []interface{}{}
	args2 = append(args2, k)
	args2 = append(args2, args...)
	return this.Execute("ZADD", args...)
}

//获取有序集合的成员数
func (this *Client) ZCard(k string) *RedisReply {
	return this.Execute("ZCARD", k)
}

//计算在有序集合中指定区间分数的成员数
func (this *Client) ZCount(k string, min, max float64) *RedisReply {
	return this.Execute("ZCOUNT", k, min, max)
}

//有序集合中对指定成员的分数加上增量 increment
func (this *Client) ZIncrBy(k string, incr float64, m string) *RedisReply {
	return this.Execute("ZINCRBY", k, incr, m)
}

//计算给定的一个或多个有序集的交集并将结果集存储在新的有序集合 destination 中,numKeys指定结果的数量
func (this *Client) ZInterStore(dest string, numkeys int, s ...string) *RedisReply {
	args := []interface{}{}
	args = append(args, dest)
	args = append(args, numkeys)
	for _, v := range s {
		args = append(args, v)
	}
	return this.Execute("ZINTERSTORE", args...)
}

//在有序集合中计算指定字典区间内成员数量
func (this *Client) ZLexCount(args ...interface{}) *RedisReply {
	return this.Execute("ZLEXCOUNT", args...)
}

//通过索引区间返回有序集合指定区间内的成员
func (this *Client) ZRange(k string, start, end int, withScore bool) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	args = append(args, start)
	args = append(args, end)
	if withScore {
		args = append(args, "WITHSCORE")
	}
	return this.Execute("ZRANGE", args...)
}

//通过字典区间返回有序集合的成员
func (this *Client) ZRangeByLex(args ...interface{}) *RedisReply {
	return this.Execute("ZRANGEBYLEX", args...)
}

//通过分数返回有序集合指定区间内的成员
func (this *Client) ZRangeByScore(k string, min, max float64, withScore bool, limit ...string) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	args = append(args, min)
	args = append(args, max)
	if withScore {
		args = append(args, "WITHSCORE")
	}
	for _, v := range limit {
		args = append(args, v)
		break
	}
	return this.Execute("ZRANGEBYSCORE", args...)
}

//返回有序集合中指定成员的索引
func (this *Client) ZRank(k string, m string) *RedisReply {
	return this.Execute("ZRANK", k, m)
}

//移除有序集合中的一个或多个成员
func (this *Client) ZRem(k string, m ...string) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	for _, v := range m {
		args = append(args, v)
		break
	}
	return this.Execute("ZREM", args...)
}

//移除有序集合中给定的字典区间的所有成员
func (this *Client) ZRemRangeByLex(args ...interface{}) *RedisReply {
	return this.Execute("ZREMRANGEBYLEX", args...)
}

//移除有序集合中给定的排名区间的所有成员
func (this *Client) ZRemRangeByRank(k string, start, end int) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	args = append(args, start)
	args = append(args, end)
	return this.Execute("ZREMRANGEBYRANK", args...)
}

//移除有序集合中给定的分数区间的所有成员
func (this *Client) ZRemRangeByScore(k string, min, max float64) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	args = append(args, min)
	args = append(args, max)
	return this.Execute("ZREMRANGEBYSCORE", args...)
}

//返回有序集中指定区间内的成员，通过索引，分数从高到低
func (this *Client) ZRevRange(k string, start, end int, withScore bool) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	args = append(args, start)
	args = append(args, end)
	if withScore {
		args = append(args, "WITHSCORE")
	}
	return this.Execute("ZREVRANGE", args...)
}

//返回有序集中指定分数区间内的成员，分数从高到低排序
func (this *Client) ZRevRangeByScore(k string, min, max float64, withScore bool) *RedisReply {
	args := []interface{}{}
	args = append(args, k)
	args = append(args, min)
	args = append(args, max)
	if withScore {
		args = append(args, "WITHSCORE")
	}
	return this.Execute("ZREVRANGEBYSCORE", args...)
}

//返回有序集合中指定成员的排名，有序集成员按分数值递减(从大到小)排序
func (this *Client) ZREVRank(k string, m string) *RedisReply {
	return this.Execute("ZREVRANK", k, m)
}

//返回有序集中，成员的分数值
func (this *Client) ZScore(k string, m string) *RedisReply {
	return this.Execute("ZSCORE", k, m)
}

//计算给定的一个或多个有序集的并集，并存储在新的 key 中
func (this *Client) ZUnionStore(dest string, numkeys int, s ...string) *RedisReply {
	args := []interface{}{}
	args = append(args, dest)
	args = append(args, numkeys)
	for _, v := range s {
		args = append(args, v)
	}
	return this.Execute("ZUNIONSTORE", args...)
}
