package dao

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/swanky2009/goim/logic/g"
	"github.com/swanky2009/goim/pkg/hash"
)

const (
	_prefixMidServer  = "mid_%d"  // mid -> key:server     hset
	_prefixKeyServer  = "key_%s"  // key -> server         string
	_prefixRoomCounts = "room_%s" // room -> server:count  hset
	_keyRooms         = "rooms"   // key -> room list 	   set
	_keyServers       = "servers" // key -> server list    sortedset
)

func keyMidServer(mid int64) string {
	return fmt.Sprintf(_prefixMidServer, mid)
}

func keyKeyServer(key string) string {
	return fmt.Sprintf(_prefixKeyServer, key)
}

func keyRoomCounts(room string) string {
	return fmt.Sprintf(_prefixRoomCounts, hash.Sha1s(room))
}

// pingRedis check redis connection.
func (d *Dao) pingRedis(c context.Context) (err error) {
	_, err = d.redis.Ping().Result()
	return
}

// AddMapping add a mapping.
// Mapping:
// mid:用户ID key:设备ID
// mid -> key_server  一个用户可同时登录多个设备
// key -> server 一个用户设备对应一个server
func (d *Dao) AddMapping(c context.Context, mid int64, key, server string) (err error) {
	if mid > 0 {
		if err = d.redis.HSet(keyMidServer(mid), key, server).Err(); err != nil {
			g.Logger.Errorf("redis.Send(HSET %d,%s,%s) error(%v)", mid, server, key, err)
			return
		}
		if err = d.redis.Expire(keyMidServer(mid), d.redisExpire).Err(); err != nil {
			g.Logger.Errorf("redis.Send(EXPIRE %d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	if err = d.redis.Set(keyKeyServer(key), server, d.redisExpire).Err(); err != nil {
		g.Logger.Errorf("redis.Send(SET %d,%s,%s) error(%v)", mid, server, key, err)
		return
	}
	return
}

// ExpireMapping expire a mapping.
func (d *Dao) ExpireMapping(c context.Context, mid int64, key string) (has bool, err error) {
	if mid > 0 {
		if has, err = d.redis.Expire(keyMidServer(mid), d.redisExpire).Result(); err != nil {
			g.Logger.Errorf("redis.Send(EXPIRE %d,%s) error(%v)", mid, key, err)
			return
		}
	}
	if has, err = d.redis.Expire(keyKeyServer(key), d.redisExpire).Result(); err != nil {
		g.Logger.Errorf("redis.Send(EXPIRE %d,%s) error(%v)", mid, key, err)
		return
	}
	return
}

// DelMapping del a mapping.
func (d *Dao) DelMapping(c context.Context, mid int64, key, server string) (has bool, err error) {
	var rows int64
	if mid > 0 {
		if rows, err = d.redis.HDel("HDEL", keyMidServer(mid), key).Result(); err != nil {
			g.Logger.Errorf("redis.Send(HDEL %d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	if rows, err = d.redis.Del(keyKeyServer(key)).Result(); err != nil {
		g.Logger.Errorf("redis.Send(DEL %d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	has = rows > 0
	return
}

// ServersByKeys get a server by key.
func (d *Dao) ServersByKeys(c context.Context, keys []string) (res []string, err error) {

	for _, key := range keys {
		args := keyKeyServer(key)

		if val, err := d.redis.Get(args).Result(); err != nil && err != redis.Nil {
			g.Logger.Errorf("redis.Do(GET %v) error(%v)", args, err)
		} else {
			res = append(res, val)
		}
	}
	return
}

// KeysByMids get a key server by mid.
func (d *Dao) KeysByMids(c context.Context, mids []int64) (ress map[string]string, olMids []int64, err error) {
	ress = make(map[string]string)
	var res map[string]string
	for i, mid := range mids {
		if res, err = d.redis.HGetAll(keyMidServer(mid)).Result(); err != nil {
			g.Logger.Errorf("redis.Do(HGETALL %d) error(%v)", mid, err)
			return
		}
		if len(res) > 0 {
			olMids = append(olMids, mids[i])
		}
		for k, v := range res {
			ress[k] = v
		}
	}
	return
}

//add server info
func (d *Dao) AddServerScore(c context.Context, server string) (err error) {
	if _, err = d.redis.ZRank(_keyServers, server).Result(); err == redis.Nil {
		if err = d.redis.ZAdd(_keyServers, redis.Z{Member: server, Score: 0}).Err(); err != nil {
			g.Logger.Errorf("redis.Do(ZAdd %s,%s) error(%v)", _keyServers, server, err)
		}
	}
	return
}

//del server info
func (d *Dao) DelServerScore(c context.Context, server string) (err error) {
	var (
		rooms []string
	)
	if err = d.redis.ZRem(_keyServers, server).Err(); err != nil {
		g.Logger.Errorf("redis.Do(ZRem %s,%s) error(%v)", _keyServers, server, err)
	}
	//del RoomCounts
	if rooms, err = d.redis.SMembers(_keyRooms).Result(); err != nil {
		g.Logger.Errorf("redis.Do(SMembers %s) error(%v)", _keyRooms, err)
		return
	}
	for _, room := range rooms {
		if err = d.redis.HDel(keyRoomCounts(room), server).Err(); err != nil {
			g.Logger.Warnf("redis.Send(HDel %s,%s) error(%v)", room, server, err)
		}
	}
	return
}

//incr server score
func (d *Dao) IncrServerScore(c context.Context, server string) (err error) {
	if err = d.redis.ZIncrBy(_keyServers, 1, server).Err(); err != nil {
		g.Logger.Errorf("redis.Do(ZIncrBy %s,%s) error(%v)", _keyServers, server, err)
	}
	return
}

//decr server score
func (d *Dao) DecrServerScore(c context.Context, server string) (err error) {
	if err = d.redis.ZIncrBy(_keyServers, -1, server).Err(); err != nil {
		g.Logger.Errorf("redis.Do(ZIncrBy %s,%s) error(%v)", _keyServers, server, err)
	}
	return
}

//get server list top
func (d *Dao) ServersRank(c context.Context, num int64) (list []string, err error) {
	if list, err = d.redis.ZRange(_keyServers, 0, num).Result(); err != nil {
		g.Logger.Errorf("redis.Do(ZRange %s) error(%v)", _keyServers, err)
	}
	return
}

func (d *Dao) UpdateRoomCount(c context.Context, server string, roomCount map[string]int32) (err error) {
	var (
		rooms []string
	)
	for room, count := range roomCount {
		if err = d.redis.SAdd(_keyRooms, room).Err(); err != nil {
			g.Logger.Warnf("redis.Send(SAdd %s,%s) error(%v)", room, err)
		}
		if err = d.redis.HSet(keyRoomCounts(room), server, count).Err(); err != nil {
			g.Logger.Warnf("redis.Send(HSet %s,%s) error(%v)", room, server, err)
		}
	}
	// server room count=0 的情况，comet不会传送，需要删除掉当前server的room count
	if rooms, err = d.redis.SMembers(_keyRooms).Result(); err != nil {
		g.Logger.Errorf("redis.Do(SMembers %s) error(%v)", _keyRooms, err)
		return
	}
	for _, room := range rooms {
		if _, ok := roomCount[room]; !ok {
			if err = d.redis.HDel(keyRoomCounts(room), server).Err(); err != nil {
				g.Logger.Warnf("redis.Send(HDel %s,%s) error(%v)", room, server, err)
			}
		}
	}
	return
}

func (d *Dao) GetAllRoomCount(c context.Context) (allroomCount map[string]int32, err error) {
	var (
		rooms    []string
		vals     []string
		count    int64
		allcount int32
	)
	if rooms, err = d.redis.SMembers(_keyRooms).Result(); err != nil {
		g.Logger.Errorf("redis.Do(SMembers %s) error(%v)", _keyRooms, err)
		return
	}

	allroomCount = make(map[string]int32, len(rooms))

	for _, room := range rooms {
		if vals, err = d.redis.HVals(keyRoomCounts(room)).Result(); err != nil {
			g.Logger.Warnf("redis.Do(HVals %s) error(%v)", room, err)
		}
		for _, val := range vals {
			count, _ = strconv.ParseInt(val, 10, 32)
			allcount += int32(count)
		}
		allroomCount[room] = allcount
	}
	return
}
