package util

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

func TryLock(redisAddr, lockKey string, ex uint, retry int) error {
	if retry <= 0 {
		retry = 10
	}
	conn, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		fmt.Println("conn to redis failed, err:%v", err)
		return err
	}
	defer conn.Close()
	ts := time.Now() // as random value
	for i := 1; i <= retry; i++ {
		if i > 1 { // sleep if not first time
			time.Sleep(time.Second)
		}
		v, err := conn.Do("SET", lockKey, ts, "EX", retry, "NX")
		if err == nil {
			if v == nil {
				fmt.Println("get lock failed, retry times:", i)
			} else {
				fmt.Println("get lock success")
				break
			}
		} else {
			fmt.Println("get lock failed with err:", err)
		}
		if i >= retry {
			err = fmt.Errorf("get lock failed with max retry times.")
			return err
		}
	}
	return nil
}

func UnLock(redisAddr, lockKey string) error {
	conn, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		fmt.Println("conn to redis failed, err:%v", err)
		return err
	}
	defer conn.Close()
	v, err := redis.Bool(conn.Do("DEL", lockKey))
	if err == nil {
		if v {
			fmt.Println("unLock success")
		} else {
			fmt.Println("unLock failed")
			return fmt.Errorf("unLock failed")
		}
	} else {
		fmt.Println("unLock failed, err:", err)
		return err
	}
	return nil
}




