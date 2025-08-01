package swarm

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
)

type fakeClient struct {
	client.Client
	infoFunc              func() (system.Info, error)
	swarmInitFunc         func(req swarm.InitRequest) (string, error)
	swarmInspectFunc      func() (swarm.Swarm, error)
	nodeInspectFunc       func() (swarm.Node, []byte, error)
	swarmGetUnlockKeyFunc func() (swarm.UnlockKeyResponse, error)
	swarmJoinFunc         func() error
	swarmLeaveFunc        func() error
	swarmUpdateFunc       func(swarm swarm.Spec, flags swarm.UpdateFlags) error
	swarmUnlockFunc       func(req swarm.UnlockRequest) error
}

func (cli *fakeClient) Info(context.Context) (system.Info, error) {
	if cli.infoFunc != nil {
		return cli.infoFunc()
	}
	return system.Info{}, nil
}

func (cli *fakeClient) NodeInspectWithRaw(context.Context, string) (swarm.Node, []byte, error) {
	if cli.nodeInspectFunc != nil {
		return cli.nodeInspectFunc()
	}
	return swarm.Node{}, []byte{}, nil
}

func (cli *fakeClient) SwarmInit(_ context.Context, req swarm.InitRequest) (string, error) {
	if cli.swarmInitFunc != nil {
		return cli.swarmInitFunc(req)
	}
	return "", nil
}

func (cli *fakeClient) SwarmInspect(context.Context) (swarm.Swarm, error) {
	if cli.swarmInspectFunc != nil {
		return cli.swarmInspectFunc()
	}
	return swarm.Swarm{}, nil
}

func (cli *fakeClient) SwarmGetUnlockKey(context.Context) (swarm.UnlockKeyResponse, error) {
	if cli.swarmGetUnlockKeyFunc != nil {
		return cli.swarmGetUnlockKeyFunc()
	}
	return swarm.UnlockKeyResponse{}, nil
}

func (cli *fakeClient) SwarmJoin(context.Context, swarm.JoinRequest) error {
	if cli.swarmJoinFunc != nil {
		return cli.swarmJoinFunc()
	}
	return nil
}

func (cli *fakeClient) SwarmLeave(context.Context, bool) error {
	if cli.swarmLeaveFunc != nil {
		return cli.swarmLeaveFunc()
	}
	return nil
}

func (cli *fakeClient) SwarmUpdate(_ context.Context, _ swarm.Version, swarmSpec swarm.Spec, flags swarm.UpdateFlags) error {
	if cli.swarmUpdateFunc != nil {
		return cli.swarmUpdateFunc(swarmSpec, flags)
	}
	return nil
}

func (cli *fakeClient) SwarmUnlock(_ context.Context, req swarm.UnlockRequest) error {
	if cli.swarmUnlockFunc != nil {
		return cli.swarmUnlockFunc(req)
	}
	return nil
}
