package main

import (
	"context"
	"log/slog"

	"github.com/zalgonoise/go-diagrams/diagram"
	"github.com/zalgonoise/go-diagrams/nodes/apps"
	"github.com/zalgonoise/go-diagrams/nodes/oci"
	"github.com/zalgonoise/x/cli"
)

//nolint:gochecknoglobals // immutable, private set of supported modes
var modes = []string{"generate"}

func main() {
	runner := cli.NewRunner("diagrams",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"generate": cli.Executable(ExecGenerate),
		}),
	)

	cli.Run(runner)
}

func ExecGenerate(_ context.Context, _ *slog.Logger, _ []string) (int, error) {
	if err := generateSDK(); err != nil {
		return 1, err
	}

	return 0, nil
}

func generateSDK() error {
	d, err := diagram.New(
		diagram.BaseDir("diagrams"),
		diagram.Direction("LR"),
		diagram.Filename("collide_api"),
		diagram.Label("Collide API"),
	)
	if err != nil {
		return err
	}

	listDistricts := oci.Database.Stream().Label("List Districts RPC")
	listAllTracksByDistrict := oci.Database.Stream().Label("List All Tracks RPC")
	getAlternativesByTrackAndDistrict := oci.Database.Stream().Label("Get Alternatives RPC")
	getCollisionsByTrackAndDistrict := oci.Database.Stream().Label("Get Collisions RPC")

	client := apps.Client.User().Label("Client")
	server := apps.Client.Client().Label("Server")
	service := oci.Database.Dis().Label("Service")
	trackList := oci.Database.Science().Label("Track List")

	logging := oci.Database.Science().Label("Logging")
	metrics := oci.Database.Science().Label("Metrics")
	tracing := oci.Database.Science().Label("Traceing")
	profiling := oci.Database.Science().Label("Profiling")

	observability := diagram.NewGroup("observability", diagram.WithBackground(diagram.BackgroundBlue)).
		Label("Observability").Add(logging, metrics, tracing, profiling)

	collide := diagram.NewGroup("collide", diagram.WithBackground(diagram.BackgroundPurple)).Label("Collide").
		Add(server)
	collide.NewGroup("repository").Label("Collide Repository").Add(trackList)
	collide.NewGroup("service").Label("Collide Service").Add(service)

	d.Add(client)
	d.Group(collide)
	d.Group(observability)

	d.Connect(client, listDistricts, diagram.Bidirectional())
	d.Connect(client, listAllTracksByDistrict, diagram.Bidirectional())
	d.Connect(client, getAlternativesByTrackAndDistrict, diagram.Bidirectional())
	d.Connect(client, getCollisionsByTrackAndDistrict, diagram.Bidirectional())

	d.Connect(listDistricts, server, diagram.Bidirectional())
	d.Connect(listAllTracksByDistrict, server, diagram.Bidirectional())
	d.Connect(getAlternativesByTrackAndDistrict, server, diagram.Bidirectional())
	d.Connect(getCollisionsByTrackAndDistrict, server, diagram.Bidirectional())

	d.Connect(server, logging)
	d.Connect(server, tracing)
	d.Connect(server, profiling)

	d.Connect(service, logging)
	d.Connect(service, tracing)
	d.Connect(service, metrics)
	d.Connect(service, profiling)

	d.Connect(trackList, logging)
	d.Connect(trackList, tracing)
	d.Connect(trackList, metrics)
	d.Connect(trackList, profiling)

	d.Connect(server, service, diagram.Bidirectional())
	d.Connect(service, trackList, diagram.Bidirectional())

	if err := d.Render(); err != nil {
		return err
	}

	return nil
}
