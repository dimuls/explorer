package pg

import (
	"context"
	"explorer"

	"github.com/doug-martin/goqu/v9"

	"google.golang.org/grpc"
)

const defaultLimit = 100

func (e *Explorer) GetPeers(ctx context.Context,
	req *explorer.GetPeersReq, opts ...grpc.CallOption) (
	*explorer.GetPeersRes, error) {

	q := e.db.From(goqu.I(peer).As("p")).
		Select(goqu.I("p").All())

	if req.ChannelId != "" {
		q = q.Join(goqu.I(peerChannel).As("pc"),
			goqu.On(goqu.Ex{"p.id": goqu.I("pc.peer_id")}))
	}

	q = q.OrderAppend(goqu.I("id").Asc())

	var ps []*explorer.Peer

	err := q.Executor().ScanStructsContext(ctx, &ps)
	if err != nil {
		return nil, err
	}

	return &explorer.GetPeersRes{
		Peers: ps,
	}, nil
}

func (e *Explorer) GetChannels(ctx context.Context,
	req *explorer.GetChannelsReq, opts ...grpc.CallOption) (
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
	req *explorer.GetChannelConfigsReq, opts ...grpc.CallOption) (
	*explorer.GetChannelConfigsRes, error) {

	q := e.db.From(goqu.I(channelConfig).As("cc")).
		Select(goqu.I("cc").All)

	if req.ChannelId != "" {
		q = q.Where(goqu.Ex{"channel_id": req.ChannelId})
	}

	q = q.OrderAppend(goqu.I("id").Desc())

	var ccs []*explorer.ChannelConfig

	err := q.Executor().ScanStructsContext(ctx, &ccs)
	if err != nil {
		return nil, err
	}

	return &explorer.GetChannelConfigsRes{
		ChannelConfigs: ccs,
	}, nil
}

func (e *Explorer) GetChaincodes(ctx context.Context,
	req *explorer.GetChaincodesReq, opts ...grpc.CallOption) (
	*explorer.GetChaincodesRes, error) {

	q := e.db.From(goqu.I(chaincode).As("c")).
		Select(goqu.I("c").All())

	where := goqu.Ex{}

	if req.PeerId != 0 {
		where["peer_id"] = req.PeerId
	}

	if req.ChannelId != "" {
		where["channel_id"] = req.ChannelId
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

func (e *Explorer) GetBlocks(ctx context.Context,
	req *explorer.GetBlocksReq, opts ...grpc.CallOption) (
	*explorer.GetBlocksRes, error) {

	q := e.db.From(goqu.I(block).As("b")).
		Select(goqu.I("b").All())

	where := goqu.Ex{}

	if req.ChannelId != "" {
		where["b.channel_id"] = req.ChannelId
	}

	if req.FromBlockId != 0 {
		where["b.id"] = goqu.Op{"lt": req.FromBlockId}
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
	req *explorer.GetTransactionsReq, opts ...grpc.CallOption) (
	*explorer.GetTransactionsRes, error) {

	q := e.db.From(goqu.I(transaction).As("t")).
		Select(goqu.I("t").All())

	where := goqu.Ex{}

	if req.ChannelId != "" {
		q = q.Join(goqu.I(block).As("b"),
			goqu.On(goqu.Ex{"t.block_id": goqu.I("b.id")}))
		where["b.channel_id"] = req.ChannelId
	}

	if req.BlockId != 0 {
		where["t.block_id"] = req.BlockId
	}

	if req.FromCreatedAt != nil {
		where["t.created_at"] = goqu.Op{"lt": req.FromCreatedAt}
	}

	q = q.Where(where).
		OrderAppend(goqu.I("created_at").Desc()).
		Limit(defaultLimit)

	var ts []*explorer.Transaction

	err := q.Executor().ScanStructsContext(ctx, &ts)
	if err != nil {
		return nil, err
	}

	return &explorer.GetTransactionsRes{
		Transactions: ts,
	}, nil
}

func (e *Explorer) GetStates(ctx context.Context,
	req *explorer.GetStatesReq, opts ...grpc.CallOption) (
	*explorer.GetStatesRes, error) {

	q := e.db.From(goqu.I(state).As("s")).
		Select(goqu.I("s").All())

	where := goqu.Ex{}

	if req.ChannelId != "" {
		q = q.Join(goqu.I(transaction).As("t"),
			goqu.On(goqu.Ex{"t.block_id": goqu.I("b.id")}))
		where["b.channel_id"] = req.ChannelId
	}

	if req.TransactionId != "" {
		where["s.transaction_id"] = req.TransactionId
	}

	q = q.Where(where).
		OrderAppend(goqu.I("key").Asc()).
		Limit(defaultLimit)

	var ss []*explorer.State

	err := q.Executor().ScanStructsContext(ctx, &ss)
	if err != nil {
		return nil, err
	}

	return &explorer.GetStatesRes{
		States: ss,
	}, nil

}

func (e *Explorer) GetOldStates(ctx context.Context,
	req *explorer.GetOldStatesReq, opts ...grpc.CallOption) (
	*explorer.GetOldStatesRes, error) {

	q := e.db.From(state).Select()

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

	var ss []*explorer.OldState

	err := q.Executor().ScanStructsContext(ctx, &ss)
	if err != nil {
		return nil, err
	}

	return &explorer.GetOldStatesRes{
		OldStates: ss,
	}, nil

}
