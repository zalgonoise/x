package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/blushft/go-diagrams/diagram"
	"github.com/blushft/go-diagrams/nodes/apps"
	"github.com/blushft/go-diagrams/nodes/oci"
	"github.com/zalgonoise/x/cli"
)

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
	d, err := diagram.New(
		diagram.Direction("LR"),
		diagram.Filename("sdk"),
		diagram.Label("Audio SDK components and workflow"),
	)
	if err != nil {
		log.Fatal(err)
	}

	source := apps.Client.Client().Label("Audio source")
	stream := oci.Database.Stream().Label("PCM byte stream")

	consumer := oci.Database.Dis().Label("Consumer")

	src := diagram.NewGroup("source").Label("Source").
		Add(source)

	processor := oci.Database.DatabaseService().Label("Processor")

	exporter := oci.Monitoring.Queue().Label("Exporter")
	emitter := oci.Network.ServiceGateway().Label("Emitter")
	collector := oci.Governance.Compartments().Label("Collector")

	output := oci.Database.Stream().Label("Output format")

	registry := oci.Governance.Ocid().Label("Registry")
	extractor := oci.Database.Science().Label("Extractor")
	compactor := oci.Storage.BlockStorage().Label("Compactor")

	expMods := diagram.NewGroup("exp_modules").Label("Exporter modules").
		Connect(emitter, collector, diagram.Bidirectional())

	colMods := diagram.NewGroup("col_modules").Label("Collector modules").
		Connect(registry, extractor, diagram.Bidirectional()).
		Connect(registry, compactor)

	core := diagram.NewGroup("sdk_core").Label("Audio SDK Core").
		Connect(consumer, processor).
		Connect(processor, exporter)

	aSDK := diagram.NewGroup("sdk").Label("Audio SDK")
	aSDK.Group(core)
	aSDK.Group(expMods)
	aSDK.Group(colMods)

	d.Group(src)
	d.Group(aSDK)
	d.Connect(exporter, collector)
	d.Connect(collector, extractor)
	d.Connect(collector, registry)
	d.Connect(source, stream).Connect(stream, consumer)
	d.Connect(emitter, output)

	if err := d.Render(); err != nil {
		return 1, err
	}

	return 0, nil
}
