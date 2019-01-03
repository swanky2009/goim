package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/swanky2009/goim/logic/g"
	"github.com/swanky2009/goim/logic/model"
	"github.com/zhenjl/cityhash"
)

const (
	_prefixMidServer    = "mid_%d"  // mid -> key:server
	_prefixKeyServer    = "key_%s"  // key -> server
	_prefixServerOnline = "ol_%s"   // server -> online
	_keyServerInfo      = "servers" // key -> server:info
)

func keyMidServer(mid int64) string {
	return fmt.Sprintf(_prefixMidServer, mid)
}

func keyKeyServer(key string) string {
	return fmt.Sprintf(_prefixKeyServer, key)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(_prefixServerOnline, key)
}

// pingRedis check redis connection.
func (d *Dao) pingRedis(c context.Context) (err error) {
	_, err = d.redis.Ping().Result()
	return
}

// AddMapping add a mapping.
// Mapping:
//	mid -> key_server
//	key -> server
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
		if err = d.redis.Expire(keyMidServer(mid), d.redisExpire).Err(); err != nil {
			g.Logger.Errorf("redis.Send(EXPIRE %d,%s) error(%v)", mid, key, err)
			return
		}
	}
	if err = d.redis.Expire(keyKeyServer(key), d.redisExpire).Err(); err != nil {
		g.Logger.Errorf("redis.Send(EXPIRE %d,%s) error(%v)", mid, key, err)
		return
	}
	return
}

// DelMapping del a mapping.
func (d *Dao) DelMapping(c context.Context, mid int64, key, server string) (has bool, err error) {

	if mid > 0 {
		if err = d.redis.HDel("HDEL", keyMidServer(mid), key).Err(); err != nil {
			g.Logger.Errorf("redis.Send(HDEL %d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	if err = d.redis.Del(keyKeyServer(key)).Err(); err != nil {
		g.Logger.Errorf("redis.Send(DEL %d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
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

// AddServerInfo add a server info.
func (d *Dao) AddServerInfo(c context.Context, server string, res *model.ServerInfo) (err error) {
	b, _ := json.Marshal(res)
	if err = d.redis.HSet(_keyServerInfo, server, b).Err(); err != nil {
		g.Logger.Errorf("redis.Send(HSET %s,%s) error(%v)", _keyServerInfo, server, err)
		return
	}
	if err = d.redis.Expire(_keyServerInfo, d.redisExpire).Err(); err != nil {
		g.Logger.Errorf("redis.Send(EXPIRE %s) error(%v)", server, err)
		return
	}
	return
}

// ServerInfos get all servers.
func (d *Dao) ServerInfos(c context.Context) (res []*model.ServerInfo, err error) {
	bs, err := d.redis.HVals(_keyServerInfo).Result()
	if err != nil {
		g.Logger.Errorf("redis.Do(GET %s) error(%v)", _keyServerInfo, err)
		return
	}
	for _, b := range bs {
		r := new(model.ServerInfo)
		if err = json.Unmarshal([]byte(b), r); err != nil {
			g.Logger.Errorf("serverOnline json.Unmarshal(%s) error(%v)", b, err)
			return
		}
		res = append(res, r)
	}
	return
}

// DelServerInfo del a server online.
func (d *Dao) DelServerInfo(c context.Context, server string) (err error) {
	if err = d.redis.HDel(_keyServerInfo, server).Err(); err != nil {
		g.Logger.Errorf("redis.Do(HDEL %s,%s) error(%v)", _keyServerInfo, server, err)
	}
	return
}

// AddServerOnline add a server online.
func (d *Dao) AddServerOnline(c context.Context, server string, online *model.Online) (err error) {
	roomsMap := map[uint32]map[string]int32{}
	for room, count := range online.RoomCount {
		rMap := roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64]
		if rMap == nil {
			rMap = make(map[string]int32)
			roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64] = rMap
		}
		rMap[room] = count
	}
	key := keyServerOnline(server)
	for hashKey, value := range roomsMap {
		err = d.addServerOnline(c, key, strconv.FormatInt(int64(hashKey), 10), &model.Online{RoomCount: value, Server: online.Server, Updated: online.Updated})
		if err != nil {
			return
		}
	}
	return
}

func (d *Dao) addServerOnline(c context.Context, key string, hashKey string, online *model.Online) (err error) {
	b, _ := json.Marshal(online)
	if err = d.redis.HSet(key, hashKey, b).Err(); err != nil {
		g.Logger.Errorf("redis.Send(HSet %s,%s) error(%v)", key, hashKey, err)
		return
	}
	if err = d.redis.Expire(key, d.redisExpire).Err(); err != nil {
		g.Logger.Errorf("redis.Send(EXPIRE %s) error(%v)", key, err)
		return
	}
	return
}

// ServerOnline get a server online.
func (d *Dao) ServerOnline(c context.Context, server string) (online *model.Online, err error) {
	online = &model.Online{RoomCount: map[string]int32{}}
	key := keyServerOnline(server)
	for i := 0; i < 64; i++ {
		ol, err := d.serverOnline(c, key, strconv.FormatInt(int64(i), 10))
		if err == nil && ol != nil {
			online.Server = ol.Server
			if ol.Updated > online.Updated {
				online.Updated = ol.Updated
			}
			for room, count := range ol.RoomCount {
				online.RoomCount[room] = count
			}
		}
	}
	return
}

func (d *Dao) serverOnline(c context.Context, key string, hashKey string) (online *model.Online, err error) {
	var res string
	res, err = d.redis.HGet(key, hashKey).Result()
	if err != nil {
		if err != redis.Nil {
			g.Logger.Errorf("redis.Do(HGET %s %s) error(%v)", key, hashKey, err)
		}
		return
	}
	online = new(model.Online)
	if err = json.Unmarshal([]byte(res), online); err != nil {
		g.Logger.Errorf("serverOnline json.Unmarshal(%s) error(%v)", res, err)
		return
	}
	return
}

// DelServerOnline del a server online.
func (d *Dao) DelServerOnline(c context.Context, server string) (err error) {
	key := keyServerOnline(server)
	if err = d.redis.Del(key).Err(); err != nil {
		g.Logger.Errorf("redis.Do(DEL %s) error(%v)", key, err)
	}
	return
}
