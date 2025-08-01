package service

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestUpdateServiceArgs(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("args", "the \"new args\"")

	spec := &swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
		},
	}
	cspec := spec.TaskTemplate.ContainerSpec
	cspec.Args = []string{"old", "args"}

	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, is.DeepEqual([]string{"the", "new args"}, cspec.Args))
}

func TestUpdateLabels(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("label-add", "add-beats-remove=value")
	flags.Set("label-add", "to-add=value")
	flags.Set("label-add", "to-update=new-value")
	flags.Set("label-add", "to-replace=new-value")
	flags.Set("label-rm", "add-beats-remove")
	flags.Set("label-rm", "to-remove")
	flags.Set("label-rm", "to-replace")
	flags.Set("label-rm", "no-such-label")

	labels := map[string]string{
		"to-keep":    "value",
		"to-remove":  "value",
		"to-replace": "value",
		"to-update":  "value",
	}

	updateLabels(flags, &labels)
	assert.DeepEqual(t, labels, map[string]string{
		"add-beats-remove": "value",
		"to-add":           "value",
		"to-keep":          "value",
		"to-replace":       "new-value",
		"to-update":        "new-value",
	})
}

func TestUpdateContainerLabels(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("container-label-add", "add-beats-remove=value")
	flags.Set("container-label-add", "to-add=value")
	flags.Set("container-label-add", "to-update=new-value")
	flags.Set("container-label-add", "to-replace=new-value")
	flags.Set("container-label-rm", "add-beats-remove")
	flags.Set("container-label-rm", "to-remove")
	flags.Set("container-label-rm", "to-replace")
	flags.Set("container-label-rm", "no-such-label")

	labels := map[string]string{
		"to-keep":    "value",
		"to-remove":  "value",
		"to-replace": "value",
		"to-update":  "value",
	}

	updateContainerLabels(flags, &labels)
	assert.DeepEqual(t, labels, map[string]string{
		"add-beats-remove": "value",
		"to-add":           "value",
		"to-keep":          "value",
		"to-replace":       "new-value",
		"to-update":        "new-value",
	})
}

func TestUpdatePlacementConstraints(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("constraint-add", "node=toadd")
	flags.Set("constraint-rm", "node!=toremove")

	placement := &swarm.Placement{
		Constraints: []string{"node!=toremove", "container=tokeep"},
	}

	updatePlacementConstraints(flags, placement)
	assert.Assert(t, is.Len(placement.Constraints, 2))
	assert.Check(t, is.Equal("container=tokeep", placement.Constraints[0]))
	assert.Check(t, is.Equal("node=toadd", placement.Constraints[1]))
}

func TestUpdatePlacementPrefs(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("placement-pref-add", "spread=node.labels.dc")
	flags.Set("placement-pref-rm", "spread=node.labels.rack")

	placement := &swarm.Placement{
		Preferences: []swarm.PlacementPreference{
			{
				Spread: &swarm.SpreadOver{
					SpreadDescriptor: "node.labels.rack",
				},
			},
			{
				Spread: &swarm.SpreadOver{
					SpreadDescriptor: "node.labels.row",
				},
			},
		},
	}

	updatePlacementPreferences(flags, placement)
	assert.Assert(t, is.Len(placement.Preferences, 2))
	assert.Check(t, is.Equal("node.labels.row", placement.Preferences[0].Spread.SpreadDescriptor))
	assert.Check(t, is.Equal("node.labels.dc", placement.Preferences[1].Spread.SpreadDescriptor))
}

func TestUpdateEnvironment(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("env-add", "toadd=newenv")
	flags.Set("env-rm", "toremove")

	envs := []string{"toremove=theenvtoremove", "tokeep=value"}

	updateEnvironment(flags, &envs)
	assert.Assert(t, is.Len(envs, 2))
	// Order has been removed in updateEnvironment (map)
	sort.Strings(envs)
	assert.Check(t, is.Equal("toadd=newenv", envs[0]))
	assert.Check(t, is.Equal("tokeep=value", envs[1]))
}

func TestUpdateEnvironmentWithDuplicateValues(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("env-rm", "foo")
	flags.Set("env-add", "foo=first")
	flags.Set("env-add", "foo=second")

	envs := []string{"foo=value"}

	updateEnvironment(flags, &envs)
	assert.Check(t, is.Len(envs, 1))
	assert.Equal(t, envs[0], "foo=second")
}

func TestUpdateEnvironmentWithDuplicateKeys(t *testing.T) {
	// Test case for #25404
	flags := newUpdateCommand(nil).Flags()
	flags.Set("env-add", "A=b")

	envs := []string{"A=c"}

	updateEnvironment(flags, &envs)
	assert.Assert(t, is.Len(envs, 1))
	assert.Check(t, is.Equal("A=b", envs[0]))
}

func TestUpdateGroups(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("group-add", "wheel")
	flags.Set("group-add", "docker")
	flags.Set("group-rm", "root")
	flags.Set("group-add", "foo")
	flags.Set("group-rm", "docker")

	groups := []string{"bar", "root"}

	updateGroups(flags, &groups)
	assert.Assert(t, is.Len(groups, 3))
	assert.Check(t, is.Equal("bar", groups[0]))
	assert.Check(t, is.Equal("foo", groups[1]))
	assert.Check(t, is.Equal("wheel", groups[2]))
}

func TestUpdateDNSConfig(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()

	// IPv4, with duplicates
	flags.Set("dns-add", "1.1.1.1")
	flags.Set("dns-add", "1.1.1.1")
	flags.Set("dns-add", "2.2.2.2")
	flags.Set("dns-rm", "3.3.3.3")
	flags.Set("dns-rm", "2.2.2.2")
	// IPv6
	flags.Set("dns-add", "2001:db8:abc8::1")
	// Invalid dns record
	assert.Check(t, is.ErrorContains(flags.Set("dns-add", "x.y.z.w"), "IP address is not correctly formatted: x.y.z.w"))

	// domains with duplicates
	flags.Set("dns-search-add", "example.com")
	flags.Set("dns-search-add", "example.com")
	flags.Set("dns-search-add", "example.org")
	flags.Set("dns-search-rm", "example.org")
	// Invalid dns search domain
	assert.ErrorContains(t, flags.Set("dns-search-add", "example$com"), "example$com is not a valid domain")

	flags.Set("dns-option-add", "ndots:9")
	flags.Set("dns-option-rm", "timeout:3")

	config := &swarm.DNSConfig{
		Nameservers: []string{"3.3.3.3", "5.5.5.5"},
		Search:      []string{"localdomain"},
		Options:     []string{"timeout:3"},
	}

	updateDNSConfig(flags, &config)

	assert.Assert(t, is.Len(config.Nameservers, 3))
	assert.Check(t, is.Equal("1.1.1.1", config.Nameservers[0]))
	assert.Check(t, is.Equal("2001:db8:abc8::1", config.Nameservers[1]))
	assert.Check(t, is.Equal("5.5.5.5", config.Nameservers[2]))

	assert.Assert(t, is.Len(config.Search, 2))
	assert.Check(t, is.Equal("example.com", config.Search[0]))
	assert.Check(t, is.Equal("localdomain", config.Search[1]))

	assert.Assert(t, is.Len(config.Options, 1))
	assert.Check(t, is.Equal(config.Options[0], "ndots:9"))
}

func TestUpdateMounts(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("mount-add", "type=volume,source=vol2,target=/toadd")
	flags.Set("mount-rm", "/toremove")

	mounts := []mount.Mount{
		{Target: "/toremove", Source: "vol1", Type: mount.TypeBind},
		{Target: "/tokeep", Source: "vol3", Type: mount.TypeBind},
	}

	updateMounts(flags, &mounts)
	assert.Assert(t, is.Len(mounts, 2))
	assert.Check(t, is.Equal("/toadd", mounts[0].Target))
	assert.Check(t, is.Equal("/tokeep", mounts[1].Target))
}

func TestUpdateMountsWithDuplicateMounts(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("mount-add", "type=volume,source=vol4,target=/toadd")

	mounts := []mount.Mount{
		{Target: "/tokeep1", Source: "vol1", Type: mount.TypeBind},
		{Target: "/toadd", Source: "vol2", Type: mount.TypeBind},
		{Target: "/tokeep2", Source: "vol3", Type: mount.TypeBind},
	}

	updateMounts(flags, &mounts)
	assert.Assert(t, is.Len(mounts, 3))
	assert.Check(t, is.Equal("/tokeep1", mounts[0].Target))
	assert.Check(t, is.Equal("/tokeep2", mounts[1].Target))
	assert.Check(t, is.Equal("/toadd", mounts[2].Target))
}

func TestUpdatePorts(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("publish-add", "1000:1000")
	flags.Set("publish-rm", "333/udp")

	portConfigs := []swarm.PortConfig{
		{TargetPort: 333, Protocol: swarm.PortConfigProtocolUDP},
		{TargetPort: 555},
	}

	err := updatePorts(flags, &portConfigs)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(portConfigs, 2))
	// Do a sort to have the order (might have changed by map)
	targetPorts := []int{int(portConfigs[0].TargetPort), int(portConfigs[1].TargetPort)}
	sort.Ints(targetPorts)
	assert.Check(t, is.Equal(555, targetPorts[0]))
	assert.Check(t, is.Equal(1000, targetPorts[1]))
}

func TestUpdatePortsDuplicate(t *testing.T) {
	// Test case for #25375
	flags := newUpdateCommand(nil).Flags()
	flags.Set("publish-add", "80:80")

	portConfigs := []swarm.PortConfig{
		{
			TargetPort:    80,
			PublishedPort: 80,
			Protocol:      swarm.PortConfigProtocolTCP,
			PublishMode:   swarm.PortConfigPublishModeIngress,
		},
	}

	err := updatePorts(flags, &portConfigs)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(portConfigs, 1))
	assert.Check(t, is.Equal(uint32(80), portConfigs[0].TargetPort))
}

func TestUpdateHealthcheckTable(t *testing.T) {
	type test struct {
		flags    [][2]string
		initial  *container.HealthConfig
		expected *container.HealthConfig
		err      string
	}
	testCases := []test{
		{
			flags:    [][2]string{{"no-healthcheck", "true"}},
			initial:  &container.HealthConfig{Test: []string{"CMD-SHELL", "cmd1"}, Retries: 10},
			expected: &container.HealthConfig{Test: []string{"NONE"}},
		},
		{
			flags:    [][2]string{{"health-cmd", "cmd1"}},
			initial:  &container.HealthConfig{Test: []string{"NONE"}},
			expected: &container.HealthConfig{Test: []string{"CMD-SHELL", "cmd1"}},
		},
		{
			flags:    [][2]string{{"health-retries", "10"}},
			initial:  &container.HealthConfig{Test: []string{"NONE"}},
			expected: &container.HealthConfig{Retries: 10},
		},
		{
			flags:    [][2]string{{"health-retries", "10"}},
			initial:  &container.HealthConfig{Test: []string{"CMD", "cmd1"}},
			expected: &container.HealthConfig{Test: []string{"CMD", "cmd1"}, Retries: 10},
		},
		{
			flags:    [][2]string{{"health-interval", "1m"}},
			initial:  &container.HealthConfig{Test: []string{"CMD", "cmd1"}},
			expected: &container.HealthConfig{Test: []string{"CMD", "cmd1"}, Interval: time.Minute},
		},
		{
			flags:    [][2]string{{"health-cmd", ""}},
			initial:  &container.HealthConfig{Test: []string{"CMD", "cmd1"}, Retries: 10},
			expected: &container.HealthConfig{Retries: 10},
		},
		{
			flags:    [][2]string{{"health-retries", "0"}},
			initial:  &container.HealthConfig{Test: []string{"CMD", "cmd1"}, Retries: 10},
			expected: &container.HealthConfig{Test: []string{"CMD", "cmd1"}},
		},
		{
			flags:    [][2]string{{"health-start-period", "1m"}},
			initial:  &container.HealthConfig{Test: []string{"CMD", "cmd1"}},
			expected: &container.HealthConfig{Test: []string{"CMD", "cmd1"}, StartPeriod: time.Minute},
		},
		{
			flags: [][2]string{{"health-cmd", "cmd1"}, {"no-healthcheck", "true"}},
			err:   "--no-healthcheck conflicts with --health-* options",
		},
		{
			flags: [][2]string{{"health-interval", "10m"}, {"no-healthcheck", "true"}},
			err:   "--no-healthcheck conflicts with --health-* options",
		},
		{
			flags: [][2]string{{"health-timeout", "1m"}, {"no-healthcheck", "true"}},
			err:   "--no-healthcheck conflicts with --health-* options",
		},
	}
	for i, c := range testCases {
		flags := newUpdateCommand(nil).Flags()
		for _, flag := range c.flags {
			flags.Set(flag[0], flag[1])
		}
		cspec := &swarm.ContainerSpec{
			Healthcheck: c.initial,
		}
		err := updateHealthcheck(flags, cspec)
		if c.err != "" {
			assert.Error(t, err, c.err)
		} else {
			assert.NilError(t, err)
			if !reflect.DeepEqual(cspec.Healthcheck, c.expected) {
				t.Errorf("incorrect result for test %d, expected health config:\n\t%#v\ngot:\n\t%#v", i, c.expected, cspec.Healthcheck)
			}
		}
	}
}

func TestUpdateHosts(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("host-add", "example.net:2.2.2.2")
	flags.Set("host-add", "ipv6.net:2001:db8:abc8::1")
	// adding the special "host-gateway" target should work
	flags.Set("host-add", "host.docker.internal:host-gateway")
	// remove with ipv6 should work
	flags.Set("host-rm", "example.net:2001:db8:abc8::1")
	// just hostname should work as well
	flags.Set("host-rm", "example.net")
	// removing the special "host-gateway" target should work
	flags.Set("host-rm", "gateway.docker.internal:host-gateway")
	// bad format error
	assert.ErrorContains(t, flags.Set("host-add", "$example.com$"), `bad format for add-host: "$example.com$"`)

	hosts := []string{"1.2.3.4 example.com", "4.3.2.1 example.org", "2001:db8:abc8::1 example.net", "gateway.docker.internal:host-gateway"}
	expected := []string{"1.2.3.4 example.com", "4.3.2.1 example.org", "2.2.2.2 example.net", "2001:db8:abc8::1 ipv6.net", "host-gateway host.docker.internal"}

	err := updateHosts(flags, &hosts)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(expected, hosts))
}

func TestUpdateHostsPreservesOrder(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("host-add", "foobar:127.0.0.2")
	flags.Set("host-add", "foobar:127.0.0.1")
	flags.Set("host-add", "foobar:127.0.0.3")

	hosts := []string{}
	err := updateHosts(flags, &hosts)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual([]string{"127.0.0.2 foobar", "127.0.0.1 foobar", "127.0.0.3 foobar"}, hosts))
}

func TestUpdateHostsReplaceEntry(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("host-add", "foobar:127.0.0.4")
	flags.Set("host-rm", "foobar:127.0.0.2")

	hosts := []string{"127.0.0.2 foobar", "127.0.0.1 foobar", "127.0.0.3 foobar"}

	err := updateHosts(flags, &hosts)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual([]string{"127.0.0.1 foobar", "127.0.0.3 foobar", "127.0.0.4 foobar"}, hosts))
}

func TestUpdateHostsRemoveHost(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("host-rm", "host1")

	hosts := []string{"127.0.0.2 host3 host1 host2 host4", "127.0.0.1 host1 host4", "127.0.0.3 host1"}

	err := updateHosts(flags, &hosts)
	assert.NilError(t, err)

	// Removing host `host1` should remove the entry from each line it appears in.
	// If there are no other hosts in the entry, the entry itself should be removed.
	assert.Check(t, is.DeepEqual([]string{"127.0.0.2 host3 host2 host4", "127.0.0.1 host4"}, hosts))
}

func TestUpdateHostsRemoveHostIP(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("host-rm", "host1:127.0.0.1")

	hosts := []string{"127.0.0.2 host3 host1 host2 host4", "127.0.0.1 host1 host4", "127.0.0.3 host1", "127.0.0.1 host1"}

	err := updateHosts(flags, &hosts)
	assert.NilError(t, err)

	// Removing host `host1` should remove the entry from each line it appears in,
	// but only if the IP-address matches. If there are no other hosts in the entry,
	// the entry itself should be removed.
	assert.Check(t, is.DeepEqual([]string{"127.0.0.2 host3 host1 host2 host4", "127.0.0.1 host4", "127.0.0.3 host1"}, hosts))
}

func TestUpdateHostsRemoveAll(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("host-add", "host-three:127.0.0.4")
	flags.Set("host-add", "host-one:127.0.0.5")
	flags.Set("host-rm", "host-one")

	hosts := []string{"127.0.0.1 host-one", "127.0.0.2 host-two", "127.0.0.3 host-one"}

	err := updateHosts(flags, &hosts)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual([]string{"127.0.0.2 host-two", "127.0.0.4 host-three", "127.0.0.5 host-one"}, hosts))
}

func TestUpdatePortsRmWithProtocol(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	flags.Set("publish-add", "8081:81")
	flags.Set("publish-add", "8082:82")
	flags.Set("publish-rm", "80")
	flags.Set("publish-rm", "81/tcp")
	flags.Set("publish-rm", "82/udp")

	portConfigs := []swarm.PortConfig{
		{
			TargetPort:    80,
			PublishedPort: 8080,
			Protocol:      swarm.PortConfigProtocolTCP,
			PublishMode:   swarm.PortConfigPublishModeIngress,
		},
	}

	err := updatePorts(flags, &portConfigs)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(portConfigs, 2))
	assert.Check(t, is.Equal(uint32(81), portConfigs[0].TargetPort))
	assert.Check(t, is.Equal(uint32(82), portConfigs[1].TargetPort))
}

type secretAPIClientMock struct {
	listResult []swarm.Secret
}

func (s secretAPIClientMock) SecretList(context.Context, swarm.SecretListOptions) ([]swarm.Secret, error) {
	return s.listResult, nil
}

func (secretAPIClientMock) SecretCreate(context.Context, swarm.SecretSpec) (swarm.SecretCreateResponse, error) {
	return swarm.SecretCreateResponse{}, nil
}

func (secretAPIClientMock) SecretRemove(context.Context, string) error {
	return nil
}

func (secretAPIClientMock) SecretInspectWithRaw(context.Context, string) (swarm.Secret, []byte, error) {
	return swarm.Secret{}, []byte{}, nil
}

func (secretAPIClientMock) SecretUpdate(context.Context, string, swarm.Version, swarm.SecretSpec) error {
	return nil
}

// TestUpdateSecretUpdateInPlace tests the ability to update the "target" of a
// secret with "docker service update" by combining "--secret-rm" and
// "--secret-add" for the same secret.
func TestUpdateSecretUpdateInPlace(t *testing.T) {
	apiClient := secretAPIClientMock{
		listResult: []swarm.Secret{
			{
				ID:   "tn9qiblgnuuut11eufquw5dev",
				Spec: swarm.SecretSpec{Annotations: swarm.Annotations{Name: "foo"}},
			},
		},
	}

	flags := newUpdateCommand(nil).Flags()
	flags.Set("secret-add", "source=foo,target=foo2")
	flags.Set("secret-rm", "foo")

	secrets := []*swarm.SecretReference{
		{
			File: &swarm.SecretReferenceFileTarget{
				Name: "foo",
				UID:  "0",
				GID:  "0",
				Mode: 292,
			},
			SecretID:   "tn9qiblgnuuut11eufquw5dev",
			SecretName: "foo",
		},
	}

	ctx := context.Background()
	updatedSecrets, err := getUpdatedSecrets(ctx, apiClient, flags, secrets)

	assert.NilError(t, err)
	assert.Assert(t, is.Len(updatedSecrets, 1))
	assert.Check(t, is.Equal("tn9qiblgnuuut11eufquw5dev", updatedSecrets[0].SecretID))
	assert.Check(t, is.Equal("foo", updatedSecrets[0].SecretName))
	assert.Check(t, is.Equal("foo2", updatedSecrets[0].File.Name))
}

func TestUpdateReadOnly(t *testing.T) {
	spec := &swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
		},
	}
	cspec := spec.TaskTemplate.ContainerSpec

	// Update with --read-only=true, changed to true
	flags := newUpdateCommand(nil).Flags()
	flags.Set("read-only", "true")
	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, cspec.ReadOnly)

	// Update without --read-only, no change
	flags = newUpdateCommand(nil).Flags()
	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, cspec.ReadOnly)

	// Update with --read-only=false, changed to false
	flags = newUpdateCommand(nil).Flags()
	flags.Set("read-only", "false")
	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, !cspec.ReadOnly)
}

func TestUpdateInit(t *testing.T) {
	spec := &swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
		},
	}
	cspec := spec.TaskTemplate.ContainerSpec

	// Update with --init=true
	flags := newUpdateCommand(nil).Flags()
	flags.Set("init", "true")
	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, is.Equal(true, *cspec.Init))

	// Update without --init, no change
	flags = newUpdateCommand(nil).Flags()
	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, is.Equal(true, *cspec.Init))

	// Update with --init=false
	flags = newUpdateCommand(nil).Flags()
	flags.Set("init", "false")
	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, is.Equal(false, *cspec.Init))
}

func TestUpdateStopSignal(t *testing.T) {
	spec := &swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
		},
	}
	cspec := spec.TaskTemplate.ContainerSpec

	// Update with --stop-signal=SIGUSR1
	flags := newUpdateCommand(nil).Flags()
	flags.Set("stop-signal", "SIGUSR1")
	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, is.Equal("SIGUSR1", cspec.StopSignal))

	// Update without --stop-signal, no change
	flags = newUpdateCommand(nil).Flags()
	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, is.Equal("SIGUSR1", cspec.StopSignal))

	// Update with --stop-signal=SIGWINCH
	flags = newUpdateCommand(nil).Flags()
	flags.Set("stop-signal", "SIGWINCH")
	updateService(context.TODO(), nil, flags, spec)
	assert.Check(t, is.Equal("SIGWINCH", cspec.StopSignal))
}

func TestUpdateIsolationValid(t *testing.T) {
	flags := newUpdateCommand(nil).Flags()
	err := flags.Set("isolation", "process")
	assert.NilError(t, err)
	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
		},
	}
	err = updateService(context.Background(), nil, flags, &spec)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(container.IsolationProcess, spec.TaskTemplate.ContainerSpec.Isolation))
}

// TestUpdateLimitsReservations tests that limits and reservations are updated,
// and that values are not updated are not reset to their default value
func TestUpdateLimitsReservations(t *testing.T) {
	// test that updating works if the service did not previously
	// have limits set (https://github.com/moby/moby/issues/38363)
	t.Run("update limits from scratch", func(t *testing.T) {
		spec := swarm.ServiceSpec{
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{},
			},
		}
		flags := newUpdateCommand(nil).Flags()
		err := flags.Set(flagLimitCPU, "2")
		assert.NilError(t, err)
		err = flags.Set(flagLimitMemory, "200M")
		assert.NilError(t, err)
		err = flags.Set(flagLimitPids, "100")
		assert.NilError(t, err)
		err = updateService(context.Background(), nil, flags, &spec)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.NanoCPUs, int64(2000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.MemoryBytes, int64(209715200)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.Pids, int64(100)))
	})

	// test that updating works if the service did not previously
	// have reservations set (https://github.com/moby/moby/issues/38363)
	t.Run("update reservations from scratch", func(t *testing.T) {
		spec := swarm.ServiceSpec{
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{},
			},
		}
		flags := newUpdateCommand(nil).Flags()
		err := flags.Set(flagReserveCPU, "2")
		assert.NilError(t, err)
		err = flags.Set(flagReserveMemory, "200M")
		assert.NilError(t, err)
		err = updateService(context.Background(), nil, flags, &spec)
		assert.NilError(t, err)
	})

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
			Resources: &swarm.ResourceRequirements{
				Limits: &swarm.Limit{
					NanoCPUs:    1000000000,
					MemoryBytes: 104857600,
					Pids:        100,
				},
				Reservations: &swarm.Resources{
					NanoCPUs:    1000000000,
					MemoryBytes: 104857600,
				},
			},
		},
	}

	// Updating without flags set should not modify existing values
	t.Run("update without flags set", func(t *testing.T) {
		flags := newUpdateCommand(nil).Flags()
		err := updateService(context.Background(), nil, flags, &spec)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.NanoCPUs, int64(1000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.MemoryBytes, int64(104857600)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.Pids, int64(100)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.NanoCPUs, int64(1000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.MemoryBytes, int64(104857600)))
	})

	// Updating CPU limit/reservation should not affect memory limit/reservation
	// and pids-limt
	t.Run("update cpu limit and reservation", func(t *testing.T) {
		flags := newUpdateCommand(nil).Flags()
		err := flags.Set(flagLimitCPU, "2")
		assert.NilError(t, err)
		err = flags.Set(flagReserveCPU, "2")
		assert.NilError(t, err)
		err = updateService(context.Background(), nil, flags, &spec)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.NanoCPUs, int64(2000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.MemoryBytes, int64(104857600)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.Pids, int64(100)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.NanoCPUs, int64(2000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.MemoryBytes, int64(104857600)))
	})

	// Updating Memory limit/reservation should not affect CPU limit/reservation
	// and pids-limt
	t.Run("update memory limit and reservation", func(t *testing.T) {
		flags := newUpdateCommand(nil).Flags()
		err := flags.Set(flagLimitMemory, "200M")
		assert.NilError(t, err)
		err = flags.Set(flagReserveMemory, "200M")
		assert.NilError(t, err)
		err = updateService(context.Background(), nil, flags, &spec)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.NanoCPUs, int64(2000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.MemoryBytes, int64(209715200)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.Pids, int64(100)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.NanoCPUs, int64(2000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.MemoryBytes, int64(209715200)))
	})

	// Updating PidsLimit should only modify PidsLimit, other values unchanged
	t.Run("update pids limit", func(t *testing.T) {
		flags := newUpdateCommand(nil).Flags()
		err := flags.Set(flagLimitPids, "2")
		assert.NilError(t, err)
		err = updateService(context.Background(), nil, flags, &spec)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.NanoCPUs, int64(2000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.MemoryBytes, int64(209715200)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.Pids, int64(2)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.NanoCPUs, int64(2000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.MemoryBytes, int64(209715200)))
	})

	t.Run("update pids limit to default", func(t *testing.T) {
		// Updating PidsLimit to 0 should work
		flags := newUpdateCommand(nil).Flags()
		err := flags.Set(flagLimitPids, "0")
		assert.NilError(t, err)
		err = updateService(context.Background(), nil, flags, &spec)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.NanoCPUs, int64(2000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.MemoryBytes, int64(209715200)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Limits.Pids, int64(0)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.NanoCPUs, int64(2000000000)))
		assert.Check(t, is.Equal(spec.TaskTemplate.Resources.Reservations.MemoryBytes, int64(209715200)))
	})
}

func TestUpdateIsolationInvalid(t *testing.T) {
	// validation depends on daemon os / version so validation should be done on the daemon side
	flags := newUpdateCommand(nil).Flags()
	err := flags.Set("isolation", "test")
	assert.NilError(t, err)
	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
		},
	}
	err = updateService(context.Background(), nil, flags, &spec)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(container.Isolation("test"), spec.TaskTemplate.ContainerSpec.Isolation))
}

func TestAddGenericResources(t *testing.T) {
	task := &swarm.TaskSpec{}
	flags := newUpdateCommand(nil).Flags()

	assert.Check(t, addGenericResources(flags, task))

	flags.Set(flagGenericResourcesAdd, "foo=1")
	assert.Check(t, addGenericResources(flags, task))
	assert.Check(t, is.Len(task.Resources.Reservations.GenericResources, 1))

	// Checks that foo isn't added a 2nd time
	flags = newUpdateCommand(nil).Flags()
	flags.Set(flagGenericResourcesAdd, "bar=1")
	assert.Check(t, addGenericResources(flags, task))
	assert.Check(t, is.Len(task.Resources.Reservations.GenericResources, 2))
}

func TestRemoveGenericResources(t *testing.T) {
	task := &swarm.TaskSpec{}
	flags := newUpdateCommand(nil).Flags()

	assert.Check(t, removeGenericResources(flags, task))

	flags.Set(flagGenericResourcesRemove, "foo")
	assert.Check(t, is.ErrorContains(removeGenericResources(flags, task), ""))

	flags = newUpdateCommand(nil).Flags()
	flags.Set(flagGenericResourcesAdd, "foo=1")
	addGenericResources(flags, task)
	flags = newUpdateCommand(nil).Flags()
	flags.Set(flagGenericResourcesAdd, "bar=1")
	addGenericResources(flags, task)

	flags = newUpdateCommand(nil).Flags()
	flags.Set(flagGenericResourcesRemove, "foo")
	assert.Check(t, removeGenericResources(flags, task))
	assert.Check(t, is.Len(task.Resources.Reservations.GenericResources, 1))
}

func TestUpdateNetworks(t *testing.T) {
	ctx := context.Background()
	nws := []network.Summary{
		{Name: "aaa-network", ID: "id555"},
		{Name: "mmm-network", ID: "id999"},
		{Name: "zzz-network", ID: "id111"},
	}

	client := &fakeClient{
		networkInspectFunc: func(ctx context.Context, networkID string, options network.InspectOptions) (network.Inspect, error) {
			for _, nw := range nws {
				if nw.ID == networkID || nw.Name == networkID {
					return nw, nil
				}
			}
			return network.Inspect{}, fmt.Errorf("network not found: %s", networkID)
		},
	}

	svc := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
			Networks: []swarm.NetworkAttachmentConfig{
				{Target: "id999"},
			},
		},
	}

	flags := newUpdateCommand(nil).Flags()
	err := flags.Set(flagNetworkAdd, "aaa-network")
	assert.NilError(t, err)
	err = updateService(ctx, client, flags, &svc)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual([]swarm.NetworkAttachmentConfig{{Target: "id555"}, {Target: "id999"}}, svc.TaskTemplate.Networks))

	flags = newUpdateCommand(nil).Flags()
	err = flags.Set(flagNetworkAdd, "aaa-network")
	assert.NilError(t, err)
	err = updateService(ctx, client, flags, &svc)
	assert.Error(t, err, "service is already attached to network aaa-network")
	assert.Check(t, is.DeepEqual([]swarm.NetworkAttachmentConfig{{Target: "id555"}, {Target: "id999"}}, svc.TaskTemplate.Networks))

	flags = newUpdateCommand(nil).Flags()
	err = flags.Set(flagNetworkAdd, "id555")
	assert.NilError(t, err)
	err = updateService(ctx, client, flags, &svc)
	assert.Error(t, err, "service is already attached to network id555")
	assert.Check(t, is.DeepEqual([]swarm.NetworkAttachmentConfig{{Target: "id555"}, {Target: "id999"}}, svc.TaskTemplate.Networks))

	flags = newUpdateCommand(nil).Flags()
	err = flags.Set(flagNetworkRemove, "id999")
	assert.NilError(t, err)
	err = updateService(ctx, client, flags, &svc)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual([]swarm.NetworkAttachmentConfig{{Target: "id555"}}, svc.TaskTemplate.Networks))

	flags = newUpdateCommand(nil).Flags()
	err = flags.Set(flagNetworkAdd, "mmm-network")
	assert.NilError(t, err)
	err = flags.Set(flagNetworkRemove, "aaa-network")
	assert.NilError(t, err)
	err = updateService(ctx, client, flags, &svc)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual([]swarm.NetworkAttachmentConfig{{Target: "id999"}}, svc.TaskTemplate.Networks))
}

func TestUpdateMaxReplicas(t *testing.T) {
	ctx := context.Background()

	svc := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{},
			Placement: &swarm.Placement{
				MaxReplicas: 1,
			},
		},
	}

	flags := newUpdateCommand(nil).Flags()
	flags.Set(flagMaxReplicas, "2")
	err := updateService(ctx, nil, flags, &svc)
	assert.NilError(t, err)

	assert.DeepEqual(t, svc.TaskTemplate.Placement, &swarm.Placement{MaxReplicas: uint64(2)})
}

func TestUpdateSysCtls(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		spec     map[string]string
		add      []string
		rm       []string
		expected map[string]string
	}{
		{
			name:     "from scratch",
			add:      []string{"sysctl.zet=value-99", "sysctl.alpha=value-1"},
			expected: map[string]string{"sysctl.zet": "value-99", "sysctl.alpha": "value-1"},
		},
		{
			name:     "append new",
			spec:     map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
			add:      []string{"new.sysctl=newvalue"},
			expected: map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2", "new.sysctl": "newvalue"},
		},
		{
			name:     "append duplicate is a no-op",
			spec:     map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
			add:      []string{"sysctl.one=value-1"},
			expected: map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
		},
		{
			name:     "remove and append existing is a no-op",
			spec:     map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
			add:      []string{"sysctl.one=value-1"},
			rm:       []string{"sysctl.one=value-1"},
			expected: map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
		},
		{
			name:     "remove and append new should append",
			spec:     map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
			add:      []string{"new.sysctl=newvalue"},
			rm:       []string{"new.sysctl=newvalue"},
			expected: map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2", "new.sysctl": "newvalue"},
		},
		{
			name:     "update existing",
			spec:     map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
			add:      []string{"sysctl.one=newvalue"},
			expected: map[string]string{"sysctl.one": "newvalue", "sysctl.two": "value-2"},
		},
		{
			name:     "update existing twice",
			spec:     map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
			add:      []string{"sysctl.one=newvalue", "sysctl.one=evennewervalue"},
			expected: map[string]string{"sysctl.one": "evennewervalue", "sysctl.two": "value-2"},
		},
		{
			name:     "remove all",
			spec:     map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
			rm:       []string{"sysctl.one=value-1", "sysctl.two=value-2"},
			expected: map[string]string{},
		},
		{
			name:     "remove by key",
			spec:     map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
			rm:       []string{"sysctl.one"},
			expected: map[string]string{"sysctl.two": "value-2"},
		},
		{
			name:     "remove by key and different value",
			spec:     map[string]string{"sysctl.one": "value-1", "sysctl.two": "value-2"},
			rm:       []string{"sysctl.one=anyvalueyoulike"},
			expected: map[string]string{"sysctl.two": "value-2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := swarm.ServiceSpec{
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: &swarm.ContainerSpec{Sysctls: tc.spec},
				},
			}
			flags := newUpdateCommand(nil).Flags()
			for _, v := range tc.add {
				assert.NilError(t, flags.Set(flagSysCtlAdd, v))
			}
			for _, v := range tc.rm {
				assert.NilError(t, flags.Set(flagSysCtlRemove, v))
			}
			err := updateService(ctx, &fakeClient{}, flags, &svc)
			assert.NilError(t, err)
			if !assert.Check(t, is.DeepEqual(svc.TaskTemplate.ContainerSpec.Sysctls, tc.expected)) {
				t.Logf("expected: %v", tc.expected)
				t.Logf("actual: %v", svc.TaskTemplate.ContainerSpec.Sysctls)
			}
		})
	}
}

func TestUpdateGetUpdatedConfigs(t *testing.T) {
	// cannedConfigs is a set of configs that we'll use over and over in the
	// tests. it's a map of Name to Config
	cannedConfigs := map[string]*swarm.Config{
		"bar": {
			ID: "barID",
			Spec: swarm.ConfigSpec{
				Annotations: swarm.Annotations{
					Name: "bar",
				},
			},
		},
		"cred": {
			ID: "credID",
			Spec: swarm.ConfigSpec{
				Annotations: swarm.Annotations{
					Name: "cred",
				},
			},
		},
		"newCred": {
			ID: "newCredID",
			Spec: swarm.ConfigSpec{
				Annotations: swarm.Annotations{
					Name: "newCred",
				},
			},
		},
	}
	// cannedConfigRefs is the same thing, but with config references instead
	// of ID, however, it just maps an arbitrary string value. this is
	// so we could have multiple config refs using the same config
	cannedConfigRefs := map[string]*swarm.ConfigReference{
		"fooRef": {
			ConfigID:   "fooID",
			ConfigName: "foo",
			File: &swarm.ConfigReferenceFileTarget{
				Name: "foo",
				UID:  "0",
				GID:  "0",
				Mode: 0o444,
			},
		},
		"barRef": {
			ConfigID:   "barID",
			ConfigName: "bar",
			File: &swarm.ConfigReferenceFileTarget{
				Name: "bar",
				UID:  "0",
				GID:  "0",
				Mode: 0o444,
			},
		},
		"bazRef": {
			ConfigID:   "bazID",
			ConfigName: "baz",
			File: &swarm.ConfigReferenceFileTarget{
				Name: "baz",
				UID:  "0",
				GID:  "0",
				Mode: 0o444,
			},
		},
		"credRef": {
			ConfigID:   "credID",
			ConfigName: "cred",
			Runtime:    &swarm.ConfigReferenceRuntimeTarget{},
		},
		"newCredRef": {
			ConfigID:   "newCredID",
			ConfigName: "newCred",
			Runtime:    &swarm.ConfigReferenceRuntimeTarget{},
		},
	}

	type flagVal [2]string
	type test struct {
		// the name of the subtest
		name string
		// flags are the flags we'll be setting
		flags []flagVal
		// oldConfigs are the configs that would already be on the service
		// it is a slice of strings corresponding to the key of
		// cannedConfigRefs
		oldConfigs []string
		// oldCredSpec is the credentialSpec being carried over from the old
		// object
		oldCredSpec *swarm.CredentialSpec
		// lookupConfigs are the configs we're expecting to be listed. it is a
		// slice of strings corresponding to the key of cannedConfigs
		lookupConfigs []string
		// expected is the configs we should get as a result. it is a slice of
		// strings corresponding to the key in cannedConfigRefs
		expected []string
	}

	testCases := []test{
		{
			name:       "no configs added or removed",
			oldConfigs: []string{"fooRef"},
			expected:   []string{"fooRef"},
		}, {
			name:          "add a config",
			flags:         []flagVal{{"config-add", "bar"}},
			oldConfigs:    []string{"fooRef"},
			lookupConfigs: []string{"bar"},
			expected:      []string{"fooRef", "barRef"},
		}, {
			name:       "remove a config",
			flags:      []flagVal{{"config-rm", "bar"}},
			oldConfigs: []string{"fooRef", "barRef"},
			expected:   []string{"fooRef"},
		}, {
			name:        "include an old credential spec",
			oldConfigs:  []string{"credRef"},
			oldCredSpec: &swarm.CredentialSpec{Config: "credID"},
			expected:    []string{"credRef"},
		}, {
			name:          "add a credential spec",
			oldConfigs:    []string{"fooRef"},
			flags:         []flagVal{{"credential-spec", "config://cred"}},
			lookupConfigs: []string{"cred"},
			expected:      []string{"fooRef", "credRef"},
		}, {
			name:          "change a credential spec",
			oldConfigs:    []string{"fooRef", "credRef"},
			oldCredSpec:   &swarm.CredentialSpec{Config: "credID"},
			flags:         []flagVal{{"credential-spec", "config://newCred"}},
			lookupConfigs: []string{"newCred"},
			expected:      []string{"fooRef", "newCredRef"},
		}, {
			name:          "credential spec no longer config",
			oldConfigs:    []string{"fooRef", "credRef"},
			oldCredSpec:   &swarm.CredentialSpec{Config: "credID"},
			flags:         []flagVal{{"credential-spec", "file://someFile"}},
			lookupConfigs: []string{},
			expected:      []string{"fooRef"},
		}, {
			name:          "credential spec becomes config",
			oldConfigs:    []string{"fooRef"},
			oldCredSpec:   &swarm.CredentialSpec{File: "someFile"},
			flags:         []flagVal{{"credential-spec", "config://cred"}},
			lookupConfigs: []string{"cred"},
			expected:      []string{"fooRef", "credRef"},
		}, {
			name:          "remove credential spec",
			oldConfigs:    []string{"fooRef", "credRef"},
			oldCredSpec:   &swarm.CredentialSpec{Config: "credID"},
			flags:         []flagVal{{"credential-spec", ""}},
			lookupConfigs: []string{},
			expected:      []string{"fooRef"},
		}, {
			name: "just frick my stuff up",
			// a more complicated test. add barRef, remove bazRef, keep fooRef,
			// change credentialSpec from credRef to newCredRef
			oldConfigs:  []string{"fooRef", "bazRef", "credRef"},
			oldCredSpec: &swarm.CredentialSpec{Config: "cred"},
			flags: []flagVal{
				{"config-add", "bar"},
				{"config-rm", "baz"},
				{"credential-spec", "config://newCred"},
			},
			lookupConfigs: []string{"bar", "newCred"},
			expected:      []string{"fooRef", "barRef", "newCredRef"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := newUpdateCommand(nil).Flags()
			for _, f := range tc.flags {
				flags.Set(f[0], f[1])
			}

			// fakeConfigAPIClientList is actually defined in create_test.go,
			// but we'll use it here as well
			var fakeClient fakeConfigAPIClientList = func(_ context.Context, opts swarm.ConfigListOptions) ([]swarm.Config, error) {
				names := opts.Filters.Get("name")
				assert.Equal(t, len(names), len(tc.lookupConfigs))

				configs := []swarm.Config{}
				for _, lookup := range tc.lookupConfigs {
					assert.Assert(t, is.Contains(names, lookup))
					cfg, ok := cannedConfigs[lookup]
					assert.Assert(t, ok)
					configs = append(configs, *cfg)
				}
				return configs, nil
			}

			// build the actual set of old configs and the container spec
			oldConfigs := []*swarm.ConfigReference{}
			for _, config := range tc.oldConfigs {
				cfg, ok := cannedConfigRefs[config]
				assert.Assert(t, ok)
				oldConfigs = append(oldConfigs, cfg)
			}

			containerSpec := &swarm.ContainerSpec{
				Configs: oldConfigs,
				Privileges: &swarm.Privileges{
					CredentialSpec: tc.oldCredSpec,
				},
			}

			ctx := context.Background()
			finalConfigs, err := getUpdatedConfigs(ctx, fakeClient, flags, containerSpec)
			assert.NilError(t, err)

			// ensure that the finalConfigs consists of all of the expected
			// configs
			assert.Equal(t, len(finalConfigs), len(tc.expected),
				"%v final configs, %v expected",
				len(finalConfigs), len(tc.expected),
			)
			for _, expected := range tc.expected {
				assert.Assert(t, is.Contains(finalConfigs, cannedConfigRefs[expected]))
			}
		})
	}
}

func TestUpdateCredSpec(t *testing.T) {
	type testCase struct {
		// name is the name of the subtest
		name string
		// flagVal is the value we're setting flagCredentialSpec to
		flagVal string
		// spec is the existing serviceSpec with its configs
		spec *swarm.ContainerSpec
		// expected is the expected value of the credential spec after the
		// function. it may be nil
		expected *swarm.CredentialSpec
	}

	testCases := []testCase{
		{
			name:     "add file credential spec",
			flagVal:  "file://somefile",
			spec:     &swarm.ContainerSpec{},
			expected: &swarm.CredentialSpec{File: "somefile"},
		}, {
			name:    "remove a file credential spec",
			flagVal: "",
			spec: &swarm.ContainerSpec{
				Privileges: &swarm.Privileges{
					CredentialSpec: &swarm.CredentialSpec{
						File: "someFile",
					},
				},
			},
			expected: nil,
		}, {
			name:     "remove when no CredentialSpec exists",
			flagVal:  "",
			spec:     &swarm.ContainerSpec{},
			expected: nil,
		}, {
			name:    "add a config credential spec",
			flagVal: "config://someConfigName",
			spec: &swarm.ContainerSpec{
				Configs: []*swarm.ConfigReference{
					{
						ConfigName: "someConfigName",
						ConfigID:   "someConfigID",
						Runtime:    &swarm.ConfigReferenceRuntimeTarget{},
					},
				},
			},
			expected: &swarm.CredentialSpec{
				Config: "someConfigID",
			},
		}, {
			name:    "remove a config credential spec",
			flagVal: "",
			spec: &swarm.ContainerSpec{
				Privileges: &swarm.Privileges{
					CredentialSpec: &swarm.CredentialSpec{
						Config: "someConfigID",
					},
				},
			},
			expected: nil,
		}, {
			name:    "update a config credential spec",
			flagVal: "config://someConfigName",
			spec: &swarm.ContainerSpec{
				Configs: []*swarm.ConfigReference{
					{
						ConfigName: "someConfigName",
						ConfigID:   "someConfigID",
						Runtime:    &swarm.ConfigReferenceRuntimeTarget{},
					},
				},
				Privileges: &swarm.Privileges{
					CredentialSpec: &swarm.CredentialSpec{
						Config: "someDifferentConfigID",
					},
				},
			},
			expected: &swarm.CredentialSpec{
				Config: "someConfigID",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flags := newUpdateCommand(nil).Flags()
			flags.Set(flagCredentialSpec, tc.flagVal)

			updateCredSpecConfig(flags, tc.spec)
			// handle the case where tc.spec.Privileges is nil
			if tc.expected == nil {
				assert.Assert(t, tc.spec.Privileges == nil || tc.spec.Privileges.CredentialSpec == nil)
				return
			}

			assert.Assert(t, tc.spec.Privileges != nil)
			assert.DeepEqual(t, tc.spec.Privileges.CredentialSpec, tc.expected)
		})
	}
}

func TestUpdateCaps(t *testing.T) {
	tests := []struct {
		// name is the name of the testcase
		name string
		// flagAdd is the value passed to --cap-add
		flagAdd []string
		// flagDrop is the value passed to --cap-drop
		flagDrop []string
		// spec is the original ContainerSpec, before being updated
		spec *swarm.ContainerSpec
		// expectedAdd is the set of requested caps the ContainerSpec should have once updated
		expectedAdd []string
		// expectedDrop is the set of dropped caps the ContainerSpec should have once updated
		expectedDrop []string
	}{
		{
			// Note that this won't be run as updateCapabilities is gated by anyChanged(flags, flagCapAdd, flagCapDrop)
			name: "Empty spec, no updates",
			spec: &swarm.ContainerSpec{},
		},
		{
			// Note that this won't be run as updateCapabilities is gated by anyChanged(flags, flagCapAdd, flagCapDrop)
			name: "No updates",
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_MOUNT", "CAP_NET_ADMIN"},
				CapabilityDrop: []string{"CAP_CHOWN", "CAP_SYS_ADMIN"},
			},
			expectedAdd:  []string{"CAP_MOUNT", "CAP_NET_ADMIN"},
			expectedDrop: []string{"CAP_CHOWN", "CAP_SYS_ADMIN"},
		},
		{
			// Note that this won't be run as updateCapabilities is gated by anyChanged(flags, flagCapAdd, flagCapDrop)
			name:     "Empty updates",
			flagAdd:  []string{},
			flagDrop: []string{},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_MOUNT", "CAP_NET_ADMIN"},
				CapabilityDrop: []string{"CAP_CHOWN", "CAP_SYS_ADMIN"},
			},
			expectedAdd:  []string{"CAP_MOUNT", "CAP_NET_ADMIN"},
			expectedDrop: []string{"CAP_CHOWN", "CAP_SYS_ADMIN"},
		},
		{
			// Note that this won't be run as updateCapabilities is gated by anyChanged(flags, flagCapAdd, flagCapDrop)
			name:     "Normalize cap-add only",
			flagAdd:  []string{},
			flagDrop: []string{},
			spec: &swarm.ContainerSpec{
				CapabilityAdd: []string{"ALL", "CAP_MOUNT", "CAP_NET_ADMIN"},
			},
			expectedAdd:  []string{"ALL"},
			expectedDrop: nil,
		},
		{
			// Note that this won't be run as updateCapabilities is gated by anyChanged(flags, flagCapAdd, flagCapDrop)
			name: "Normalize cap-drop only",
			spec: &swarm.ContainerSpec{
				CapabilityDrop: []string{"ALL", "CAP_MOUNT", "CAP_NET_ADMIN"},
			},
			expectedDrop: []string{"ALL"},
		},
		{
			name:         "Add new caps",
			flagAdd:      []string{"CAP_NET_ADMIN"},
			flagDrop:     []string{},
			spec:         &swarm.ContainerSpec{},
			expectedAdd:  []string{"CAP_NET_ADMIN"},
			expectedDrop: nil,
		},
		{
			name:         "Drop new caps",
			flagAdd:      []string{},
			flagDrop:     []string{"CAP_NET_ADMIN"},
			spec:         &swarm.ContainerSpec{},
			expectedAdd:  nil,
			expectedDrop: []string{"CAP_NET_ADMIN"},
		},
		{
			name:     "Add a previously dropped cap",
			flagAdd:  []string{"CAP_NET_ADMIN"},
			flagDrop: []string{},
			spec: &swarm.ContainerSpec{
				CapabilityDrop: []string{"CAP_NET_ADMIN"},
			},
			expectedAdd:  nil,
			expectedDrop: nil,
		},
		{
			name:     "Drop a previously requested cap, and add a new one",
			flagAdd:  []string{"CAP_CHOWN"},
			flagDrop: []string{"CAP_NET_ADMIN"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd: []string{"CAP_NET_ADMIN"},
			},
			expectedAdd:  []string{"CAP_CHOWN"},
			expectedDrop: nil,
		},
		{
			name:    "Add caps to service that has ALL caps has no effect",
			flagAdd: []string{"CAP_NET_ADMIN"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd: []string{"ALL"},
			},
			expectedAdd:  []string{"ALL"},
			expectedDrop: nil,
		},
		{
			name:     "Drop ALL caps, then add new caps to service that has ALL caps",
			flagAdd:  []string{"CAP_NET_ADMIN"},
			flagDrop: []string{"ALL"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd: []string{"ALL"},
			},
			expectedAdd:  []string{"CAP_NET_ADMIN"},
			expectedDrop: nil,
		},
		{
			name:         "Add takes precedence on empty spec",
			flagAdd:      []string{"CAP_NET_ADMIN"},
			flagDrop:     []string{"CAP_NET_ADMIN"},
			spec:         &swarm.ContainerSpec{},
			expectedAdd:  []string{"CAP_NET_ADMIN"},
			expectedDrop: nil,
		},
		{
			name:     "Add takes precedence on existing spec",
			flagAdd:  []string{"CAP_NET_ADMIN"},
			flagDrop: []string{"CAP_NET_ADMIN"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_NET_ADMIN"},
				CapabilityDrop: []string{"CAP_NET_ADMIN"},
			},
			expectedAdd:  []string{"CAP_NET_ADMIN"},
			expectedDrop: nil,
		},
		{
			name:     "Drop all, and add new caps",
			flagAdd:  []string{"CAP_CHOWN"},
			flagDrop: []string{"ALL"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_NET_ADMIN", "CAP_MOUNT"},
				CapabilityDrop: []string{"CAP_NET_ADMIN", "CAP_MOUNT"},
			},
			expectedAdd:  []string{"CAP_CHOWN", "CAP_MOUNT", "CAP_NET_ADMIN"},
			expectedDrop: []string{"ALL"},
		},
		{
			name:     "Add all caps",
			flagAdd:  []string{"ALL"},
			flagDrop: []string{"CAP_NET_ADMIN", "CAP_SYS_ADMIN"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_NET_ADMIN"},
				CapabilityDrop: []string{"CAP_CHOWN"},
			},
			expectedAdd:  []string{"ALL"},
			expectedDrop: []string{"CAP_CHOWN", "CAP_SYS_ADMIN"},
		},
		{
			name:     "Drop all, and add all",
			flagAdd:  []string{"ALL"},
			flagDrop: []string{"ALL"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_NET_ADMIN"},
				CapabilityDrop: []string{"CAP_CHOWN"},
			},
			expectedAdd:  []string{"ALL"},
			expectedDrop: []string{"CAP_CHOWN"},
		},
		{
			name:     "Caps are normalized and sorted",
			flagAdd:  []string{"bbb", "aaa", "cAp_bBb", "cAp_aAa"},
			flagDrop: []string{"zzz", "yyy", "cAp_yYy", "cAp_yYy"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"ccc", "CAP_DDD"},
				CapabilityDrop: []string{"www", "CAP_XXX"},
			},
			expectedAdd:  []string{"CAP_AAA", "CAP_BBB", "CAP_CCC", "CAP_DDD"},
			expectedDrop: []string{"CAP_WWW", "CAP_XXX", "CAP_YYY", "CAP_ZZZ"},
		},
		{
			name:     "Reset capabilities",
			flagAdd:  []string{"RESET"},
			flagDrop: []string{"RESET"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_AAA", "CAP_BBB", "CAP_CCC", "CAP_DDD"},
				CapabilityDrop: []string{"CAP_WWW", "CAP_XXX", "CAP_YYY", "CAP_ZZZ"},
			},
			expectedAdd:  nil,
			expectedDrop: nil,
		},
		{
			name:     "Reset capabilities, and update after",
			flagAdd:  []string{"RESET", "CAP_ADD_ONE", "CAP_FOO"},
			flagDrop: []string{"RESET", "CAP_DROP_ONE", "CAP_FOO"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_AAA", "CAP_BBB", "CAP_CCC", "CAP_DDD"},
				CapabilityDrop: []string{"CAP_WWW", "CAP_XXX", "CAP_YYY", "CAP_ZZZ"},
			},
			expectedAdd:  []string{"CAP_ADD_ONE", "CAP_FOO"},
			expectedDrop: []string{"CAP_DROP_ONE"},
		},
		{
			name:     "Reset capabilities, and add ALL",
			flagAdd:  []string{"RESET", "ALL"},
			flagDrop: []string{"RESET", "ALL"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_AAA", "CAP_BBB", "CAP_CCC", "CAP_DDD"},
				CapabilityDrop: []string{"CAP_WWW", "CAP_XXX", "CAP_YYY", "CAP_ZZZ"},
			},
			expectedAdd:  []string{"ALL"},
			expectedDrop: nil,
		},
		{
			name:     "Add ALL and RESET",
			flagAdd:  []string{"ALL", "RESET"},
			flagDrop: []string{"ALL", "RESET"},
			spec: &swarm.ContainerSpec{
				CapabilityAdd:  []string{"CAP_AAA", "CAP_BBB", "CAP_CCC", "CAP_DDD"},
				CapabilityDrop: []string{"CAP_WWW", "CAP_XXX", "CAP_YYY", "CAP_ZZZ"},
			},
			expectedAdd:  []string{"ALL"},
			expectedDrop: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := newUpdateCommand(nil).Flags()
			for _, c := range tc.flagAdd {
				_ = flags.Set(flagCapAdd, c)
			}
			for _, c := range tc.flagDrop {
				_ = flags.Set(flagCapDrop, c)
			}

			updateCapabilities(flags, tc.spec)

			assert.DeepEqual(t, tc.spec.CapabilityAdd, tc.expectedAdd)
			assert.DeepEqual(t, tc.spec.CapabilityDrop, tc.expectedDrop)
		})
	}
}

func TestUpdateUlimits(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		spec     []*container.Ulimit
		rm       []string
		add      []string
		expected []*container.Ulimit
	}{
		{
			name: "from scratch",
			add:  []string{"nofile=512:1024", "core=1024:1024"},
			expected: []*container.Ulimit{
				{Name: "core", Hard: 1024, Soft: 1024},
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
		},
		{
			name: "append new",
			spec: []*container.Ulimit{
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
			add: []string{"core=1024:1024"},
			expected: []*container.Ulimit{
				{Name: "core", Hard: 1024, Soft: 1024},
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
		},
		{
			name: "remove and append new should append",
			spec: []*container.Ulimit{
				{Name: "core", Hard: 1024, Soft: 1024},
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
			rm:  []string{"nofile=512:1024"},
			add: []string{"nofile=512:1024"},
			expected: []*container.Ulimit{
				{Name: "core", Hard: 1024, Soft: 1024},
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
		},
		{
			name: "update existing",
			spec: []*container.Ulimit{
				{Name: "nofile", Hard: 2048, Soft: 1024},
			},
			add: []string{"nofile=512:1024"},
			expected: []*container.Ulimit{
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
		},
		{
			name: "update existing twice",
			spec: []*container.Ulimit{
				{Name: "nofile", Hard: 2048, Soft: 1024},
			},
			add: []string{"nofile=256:512", "nofile=512:1024"},
			expected: []*container.Ulimit{
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
		},
		{
			name: "remove all",
			spec: []*container.Ulimit{
				{Name: "core", Hard: 1024, Soft: 1024},
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
			rm:       []string{"nofile=512:1024", "core=1024:1024"},
			expected: nil,
		},
		{
			name: "remove by key",
			spec: []*container.Ulimit{
				{Name: "core", Hard: 1024, Soft: 1024},
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
			rm: []string{"core"},
			expected: []*container.Ulimit{
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
		},
		{
			name: "remove by key and different value",
			spec: []*container.Ulimit{
				{Name: "core", Hard: 1024, Soft: 1024},
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
			rm: []string{"core=1234:5678"},
			expected: []*container.Ulimit{
				{Name: "nofile", Hard: 1024, Soft: 512},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := swarm.ServiceSpec{
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: &swarm.ContainerSpec{Ulimits: tc.spec},
				},
			}
			flags := newUpdateCommand(nil).Flags()
			for _, v := range tc.add {
				assert.NilError(t, flags.Set(flagUlimitAdd, v))
			}
			for _, v := range tc.rm {
				assert.NilError(t, flags.Set(flagUlimitRemove, v))
			}
			err := updateService(ctx, &fakeClient{}, flags, &svc)
			assert.NilError(t, err)
			assert.DeepEqual(t, svc.TaskTemplate.ContainerSpec.Ulimits, tc.expected)
		})
	}
}
