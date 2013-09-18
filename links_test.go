package docker

import (
	"fmt"
	"github.com/dotcloud/docker/utils"
	"strings"
	"testing"
)

func newTestLinkRepository(t *testing.T) *LinkRepository {
	r, err := NewLinkRepository("")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func newMockLinkContainer(id string, ip string) *Container {
	return &Container{
		Config: &Config{},
		ID:     id,
		NetworkSettings: &NetworkSettings{
			IPAddress: ip,
		},
	}
}

func TestLinkNew(t *testing.T) {
	r := newTestLinkRepository(t)
	toID := GenerateID()
	fromID := GenerateID()

	from := newMockLinkContainer(fromID, "172.0.17.2")
	from.Config.Env = []string{}
	from.State = State{Running: true}
	ports := make(map[Port]struct{})

	ports[Port("6379/tcp")] = struct{}{}

	from.Config.ExposedPorts = ports

	to := newMockLinkContainer(toID, "172.0.17.3")

	link, err := r.NewLink(to, from, "172.0.17.1", "docker")
	if err != nil {
		t.Fatal(err)
	}

	if link == nil {
		t.FailNow()
	}
	if link.ID() != fmt.Sprintf("%s:%s", utils.TruncateID(to.ID), "DOCKER") {
		t.Fail()
	}
	if link.Alias != "DOCKER" {
		t.Fail()
	}
	if link.FromID != utils.TruncateID(from.ID) {
		t.Fail()
	}
	if link.ToID != utils.TruncateID(to.ID) {
		t.Fail()
	}
	if link.ToIP != "172.0.17.3" {
		t.Fail()
	}
	if link.FromIP != "172.0.17.2" {
		t.Fail()
	}
	if link.BridgeInterface != "172.0.17.1" {
		t.Fail()
	}
	for _, p := range link.ports {
		if p != Port("6379/tcp") {
			t.Fail()
		}
	}
}

func TestLinkEnv(t *testing.T) {
	r := newTestLinkRepository(t)
	toID := GenerateID()
	fromID := GenerateID()

	from := newMockLinkContainer(fromID, "172.0.17.2")
	from.Config.Env = []string{"PASSWORD=gordon"}
	from.State = State{Running: true}
	ports := make(map[Port]struct{})

	ports[Port("6379/tcp")] = struct{}{}

	from.Config.ExposedPorts = ports

	to := newMockLinkContainer(toID, "172.0.17.3")

	link, err := r.NewLink(to, from, "172.0.17.1", "docker")
	if err != nil {
		t.Fatal(err)
	}

	rawEnv := link.ToEnv()
	env := make(map[string]string, len(rawEnv))
	for _, e := range rawEnv {
		parts := strings.Split(e, "=")
		if len(parts) != 2 {
			t.FailNow()
		}
		env[parts[0]] = parts[1]
	}
	if env["DOCKER_PORT"] != "tcp://172.0.17.2:6379" {
		t.Fail()
	}
	if env["DOCKER_PORT_6379_TCP"] != "tcp://172.0.17.2:6379" {
		t.Fail()
	}
	if env["DOCKER_ID"] != utils.TruncateID(from.ID) {
		t.Fail()
	}
	if env["DOCKER_ENV_PASSWORD"] != "gordon" {
		t.Fail()
	}
}
