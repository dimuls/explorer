package pg

import (
	"context"
	"explorer"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/doug-martin/goqu/v9"
)

const defaultLimit = 100

func (e *Explorer) PostLogin(ctx context.Context, in *explorer.PostLoginReq) (
	*explorer.PostLoginRes, error) {
	return nil, fmt.Errorf("TODO")
}

func (e *Explorer) GetPeers(ctx context.Context, req *explorer.GetPeersReq) (
	*explorer.GetPeersRes, error) {

	q := e.db.From(goqu.I(peer).As("p")).
		Select(goqu.I("p.id").As("id"),
			goqu.I("p.url").As("url")).
		Order(goqu.I("p.id").Asc())

	if req.ChannelId != 0 {
		q = q.Join(goqu.I(peerChannel).As("pc"),
			goqu.On(goqu.Ex{"p.id": goqu.I("pc.peer_id")})).
			Where(goqu.Ex{"pc.channel_id": req.ChannelId})
	}

	var ps []*explorer.Peer

	err := q.ScanStructsContext(ctx, &ps)
	if err != nil {
		return nil, err
	}

	return &explorer.GetPeersRes{
		Peers: ps,
	}, nil
}

func (e *Explorer) GetChannels(ctx context.Context,
	req *explorer.GetChannelsReq) (
	*explorer.GetChannelsRes, error) {

	q := e.db.From(goqu.I(channel).As("c")).
		Select(goqu.I("c").All())

	if req.PeerId != 0 {
		q = q.Join(goqu.I(peerChannel).As("pc"),
			goqu.On(goqu.Ex{"c.id": goqu.I("pc.channel_id")})).
			Where(goqu.Ex{"pc.peer_id": req.PeerId})
	}

	q = q.OrderAppend(goqu.I("id").Asc())

	var ss []*explorer.Channel

	err := q.Executor().ScanStructsContext(ctx, &ss)
	if err != nil {
		return nil, err
	}

	return &explorer.GetChannelsRes{
		Channels: ss,
	}, nil
}

func (e *Explorer) GetChannelConfigs(ctx context.Context,
	req *explorer.GetChannelConfigsReq) (
	*explorer.GetChannelConfigsRes, error) {

	q := e.db.From(goqu.I(channelConfig).As("cc")).
		Select("cc.id", "cc.channel_id", "cc.raw", "cc.parsed",
			"cc.created_at")

	if req.ChannelId != 0 {
		q = q.Where(goqu.Ex{"channel_id": req.ChannelId})
	}

	q = q.OrderAppend(goqu.I("id").Desc())

	var ccs []*explorer.ChannelConfig

	rows, err := q.Executor().QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		cc := &explorer.ChannelConfig{}
		var createdAt time.Time
		err = rows.Scan(&cc.Id, &cc.ChannelId, &cc.Raw, &cc.Parsed, &createdAt)
		if err != nil {
			return nil, err
		}
		cc.CreatedAt = timestamppb.New(createdAt)
		ccs = append(ccs, cc)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &explorer.GetChannelConfigsRes{
		ChannelConfigs: ccs,
	}, nil
}

func (e *Explorer) GetChaincodes(ctx context.Context,
	req *explorer.GetChaincodesReq) (
	*explorer.GetChaincodesRes, error) {

	q := e.db.From(goqu.I(chaincode).As("c")).
		Select(goqu.I("c.id").As("id"),
			goqu.I("c.name").As("name"),
			goqu.I("c.version").As("version"))

	where := goqu.Ex{}

	var ccJoined bool

	if req.PeerId != 0 {
		q = q.Join(goqu.I(channelChaincode).As("cc"),
			goqu.On(goqu.Ex{"c.id": goqu.I("cc.chaincode_id")}))
		ccJoined = true
		q = q.Join(goqu.I(peerChannel).As("pc"),
			goqu.On(goqu.Ex{"cc.channel_id": goqu.I("pc.channel_id")}))
		where["pc.peer_id"] = req.PeerId
	}

	if req.ChannelId != 0 {
		if !ccJoined {
			q = q.Join(goqu.I(channelChaincode).As("cc"),
				goqu.On(goqu.Ex{"c.id": goqu.I("cc.chaincode_id")}))
			ccJoined = true
		}
		where["cc.channel_id"] = req.ChannelId
	}

	q = q.Where(where).OrderAppend(goqu.I("id").Asc())

	var cs []*explorer.Chaincode

	err := q.Executor().ScanStructsContext(ctx, &cs)
	if err != nil {
		return nil, err
	}

	return &explorer.GetChaincodesRes{
		Chaincodes: cs,
	}, nil
}

type blockLoader struct {
	BlockID int64 `json:"block_id"`
}

func (e *Explorer) GetBlocks(ctx context.Context, req *explorer.GetBlocksReq) (
	*explorer.GetBlocksRes, error) {

	q := e.db.From(goqu.I(block).As("b")).
		Select(goqu.I("b").All())

	where := goqu.Ex{}

	if req.ChannelId != 0 {
		where["b.channel_id"] = req.ChannelId
	}

	if req.FromId != 0 {
		where["b.id"] = goqu.Op{"lt": req.FromId}
	}

	q = q.Where(where).
		OrderAppend(goqu.I("id").Desc()).
		Limit(defaultLimit)

	var bs []*explorer.Block

	err := q.Executor().ScanStructsContext(ctx, &bs)
	if err != nil {
		return nil, err
	}

	return &explorer.GetBlocksRes{
		Blocks: bs,
	}, nil

}

func (e *Explorer) GetTransactions(ctx context.Context,
	req *explorer.GetTransactionsReq) (
	*explorer.GetTransactionsRes, error) {

	q := e.db.From(transaction).
		Select("id", "channel_id", "block_id", "created_at")

	where := goqu.Ex{}

	if req.ChannelId != 0 {
		where["channel_id"] = req.ChannelId
	}

	if req.BlockId != 0 {
		where["block_id"] = req.BlockId
	}

	if req.FromCreatedAt != nil {
		op := "lte"
		if req.LoadMore {
			op = "lt"
		}
		where["created_at"] = goqu.Op{op: req.FromCreatedAt.AsTime()}
	}

	q = q.Where(where).
		OrderAppend(goqu.I("created_at").Desc()).
		Limit(defaultLimit)

	rows, err := q.Executor().QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ts []*explorer.Transaction

	for rows.Next() {
		t := &explorer.Transaction{}
		var createdAt time.Time
		err = rows.Scan(&t.Id, &t.ChannelId, &t.BlockId, &createdAt)
		if err != nil {
			return nil, err
		}
		t.CreatedAt = timestamppb.New(createdAt)
		ts = append(ts, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &explorer.GetTransactionsRes{
		Transactions: ts,
	}, nil
}

func (e *Explorer) GetStates(ctx context.Context, req *explorer.GetStatesReq) (
	*explorer.GetStatesRes, error) {

	q := e.db.From(goqu.I(state).As("s")).
		Select("s.key", "s.channel_id", "s.transaction_id", "s.type",
			"s.raw_value", "s.value", "s.created_at")

	where := goqu.Ex{}

	if req.ChannelId != 0 {
		where["s.channel_id"] = req.ChannelId
	}

	if req.TransactionId != "" {
		where["s.transaction_id"] = req.TransactionId
	}

	if req.FromCreatedAt != nil {
		op := "lte"
		if req.LoadMore {
			op = "lt"
		}
		where["s.created_at"] = goqu.Op{op: req.FromCreatedAt.AsTime()}
	}

	q = q.Where(where).
		OrderAppend(goqu.I("key").Asc()).
		Limit(defaultLimit)

	fmt.Println(q.ToSQL())

	rows, err := q.Executor().QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ss []*explorer.State

	for rows.Next() {
		s := &explorer.State{}
		var createdAt time.Time
		err = rows.Scan(&s.Key, &s.ChannelId, &s.TransactionId, &s.Type,
			&s.RawValue, &s.Value, &createdAt)
		if err != nil {
			return nil, err
		}
		s.CreatedAt = timestamppb.New(createdAt)
		ss = append(ss, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &explorer.GetStatesRes{
		States: ss,
	}, nil

}

func (e *Explorer) GetOldStates(ctx context.Context,
	req *explorer.GetOldStatesReq) (
	*explorer.GetOldStatesRes, error) {

	q := e.db.From(state).Select("id", "channel_id",
		"transaction_id", "key", "type", "raw_value", "value", "created_at")

	where := goqu.Ex{}

	if req.Key != "" {
		where["key"] = req.Key
	}

	if req.FromId != 0 {
		where["id"] = goqu.Op{"lt": req.FromId}
	}

	q = q.Where(where).
		OrderAppend(goqu.I("id").Desc()).
		Limit(defaultLimit)

	rows, err := q.Executor().QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ss []*explorer.OldState

	for rows.Next() {
		s := &explorer.OldState{}
		var createdAt time.Time
		err = rows.Scan(&s.Id, &s.ChannelId, &s.TransactionId, &s.Key, &s.Type,
			&s.RawValue, &s.Value, &createdAt)
		if err != nil {
			return nil, err
		}
		s.CreatedAt = timestamppb.New(createdAt)
		ss = append(ss, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &explorer.GetOldStatesRes{
		OldStates: ss,
	}, nil

}
