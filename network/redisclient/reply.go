package redisclient

import "github.com/gomodule/redigo/redis"

type RedisReply struct {
	Value interface{}
	Error error
}

func NewReply(v interface{},e error)*RedisReply{
	return &RedisReply{
		Value: v,
		Error: e,
	}
}


func (this *RedisReply) Interface()(interface{},error) {
	return this.Value,this.Error
}

func (this *RedisReply) Int()(int,error) {
	return redis.Int(this.Value,this.Error)
}

func (this *RedisReply) Int64()(int64,error) {
	return redis.Int64(this.Value,this.Error)
}

func (this *RedisReply) Uint64()(uint64,error) {
	return redis.Uint64(this.Value,this.Error)
}

func (this *RedisReply) Float64()(float64,error) {
	return redis.Float64(this.Value,this.Error)
}

func (this *RedisReply) String()(string,error) {
	return redis.String(this.Value,this.Error)
}

func (this *RedisReply) Bytes()([]byte,error) {
	return redis.Bytes(this.Value,this.Error)
}

func (this *RedisReply) Bool()(bool,error) {
	return redis.Bool(this.Value,this.Error)
}

func (this *RedisReply) MultiBulk()([]interface{},error) {
	return redis.MultiBulk(this.Value,this.Error)
}

func (this *RedisReply) Values()([]interface{},error) {
	return redis.Values(this.Value,this.Error)
}

func (this *RedisReply) Float64s()([]float64,error) {
	return redis.Float64s(this.Value,this.Error)
}

func (this *RedisReply) Strings()([]string,error) {
	return redis.Strings(this.Value,this.Error)
}

func (this *RedisReply) ByteSlices()([][]byte,error) {
	return redis.ByteSlices(this.Value,this.Error)
}

func (this *RedisReply) Int64s()([]int64,error) {
	return redis.Int64s(this.Value,this.Error)
}


func (this *RedisReply) Ints()([]int,error) {
	return redis.Ints(this.Value,this.Error)
}


func (this *RedisReply) StringMap()(map[string]string,error) {
	return redis.StringMap(this.Value,this.Error)
}


func (this *RedisReply) IntMap()(map[string]int,error) {
	return redis.IntMap(this.Value,this.Error)
}


func (this *RedisReply) Int64Map()(map[string]int64,error) {
	return redis.Int64Map(this.Value,this.Error)
}


func (this *RedisReply) Positions()([]*[2]float64,error) {
	return redis.Positions(this.Value,this.Error)
}


func (this *RedisReply) Uint64s()([]uint64,error) {
	return redis.Uint64s(this.Value,this.Error)
}


func (this *RedisReply) Uint64Map()(map[string]uint64,error) {
	return redis.Uint64Map(this.Value,this.Error)
}

func (this *RedisReply) SlowLogs()([]redis.SlowLog,error) {
	return redis.SlowLogs(this.Value,this.Error)
}


