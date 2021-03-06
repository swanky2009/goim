package job

import (
	"context"
	"fmt"

	pb_c "github.com/swanky2009/goim/grpc/comet"
	pb_l "github.com/swanky2009/goim/grpc/logic"
	"github.com/swanky2009/goim/job/g"
)

func (j *Job) push(ctx context.Context, m *pb_l.PushMsg) (err error) {
	switch m.Type {
	case pb_l.PushMsg_PUSH:

		proto := &pb_c.Proto{Ver: 0, Op: m.Operation, Body: m.Msg}

		j.comets.Push(m.Server, &pb_c.PushMsgReq{Keys: m.Keys, ProtoOp: m.Operation, Proto: proto})

		g.Logger.Debugf("push msg serverId: %s keys(%v)", m.Server, m.Keys)

	case pb_l.PushMsg_ROOM:

		proto := &pb_c.Proto{Ver: 0, Op: m.Operation, Body: m.Msg}

		j.comets.BroadcastRoom(m.Room, &pb_c.BroadcastRoomReq{RoomID: m.Room, Proto: proto})

	case pb_l.PushMsg_BROADCAST:

		proto := &pb_c.Proto{Ver: 0, Op: m.Operation, Body: m.Msg}

		j.comets.Broadcast(&pb_c.BroadcastReq{ProtoOp: m.Operation, Proto: proto, Speed: m.Speed, Platform: m.Platform})

	default:
		err = fmt.Errorf("no match type: %s", m.Type)
	}
	return
}
