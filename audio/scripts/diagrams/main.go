package main

import (
	"context"
	"log/slog"

	"github.com/zalgonoise/go-diagrams/diagram"
	"github.com/zalgonoise/go-diagrams/nodes/apps"
	"github.com/zalgonoise/go-diagrams/nodes/oci"
	"github.com/zalgonoise/x/cli"
)

const modeGenerate = "generate"

//nolint:gochecknoglobals // immutable, private set of supported modes
var modes = []string{modeGenerate}

func main() {
	runner := cli.NewRunner("diagrams",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			modeGenerate: cli.Executable(ExecGenerate),
		}),
	)

	cli.Run(runner)
}

func ExecGenerate(_ context.Context, _ *slog.Logger, _ []string) (int, error) {
	if err := generateSDK(); err != nil {
		return 1, err
	}

	if err := generateEncoding(); err != nil {
		return 1, err
	}

	return 0, nil
}

func generateSDK() error {
	d, err := diagram.New(
		diagram.BaseDir("diagrams"),
		diagram.Direction("LR"),
		diagram.Filename("sdk"),
		diagram.Label("Audio SDK: the pipeline (data plane) and metrics (analysis plane)"),
	)
	if err != nil {
		return err
	}

	client := apps.Client.Client().Label("Audio source")
	file := oci.Storage.FileStorage().Label("File")
	reader := oci.Database.Stream().Label("Reader (stdin)")
	httpSource := oci.Database.Stream().Label("HTTP")

	decode := oci.Database.Dis().Label("audio.Decode")
	transforms := oci.Database.Science().Label("Transforms")
	exporters := oci.Network.ServiceGateway().Label("Exporters, in order")

	dropQueue := oci.Monitoring.Queue().Label("QueuedExporter (drop oldest)")
	blockQueue := oci.Monitoring.Queue().Label("QueuedExporter (block)")
	wavExporter := oci.Storage.BlockStorage().Label("WAV exporter")

	meter := oci.Governance.Compartments().Label("Meter")
	instruments := oci.Database.Science().Label("Instruments (levels, spectra)")
	aggregations := oci.Governance.Ocid().Label("Aggregations")
	periodic := oci.Monitoring.Events().Label("PeriodicReader")
	manual := oci.Monitoring.Events().Label("ManualReader")
	slogEmitter := oci.Database.Stream().Label("SlogEmitter (JSON lines)")
	prometheus := oci.Monitoring.Telemetry().Label("Prometheus Collector")

	recorder := oci.Storage.BlockStorage().Label("Recorder (segments)")
	dir := oci.Storage.FileStorage().Label("Dir store")

	sources := diagram.NewGroup("sources").Label("sdk/sources").Add(file, reader, httpSource)

	pipeline := diagram.NewGroup("pipeline").Label("sdk.Pipeline").
		Connect(decode, transforms).
		Connect(transforms, exporters)

	sdk := diagram.NewGroup("exporters").Label("Exporters")
	sdk.NewGroup("queues", diagram.WithBackground(diagram.BackgroundYellow)).
		Label("sdk").
		Add(dropQueue, blockQueue)
	sdk.NewGroup("wav", diagram.WithBackground(diagram.BackgroundYellow)).
		Label("sdk/exporters").
		Add(wavExporter)
	sdk.NewGroup("metric", diagram.WithBackground(diagram.BackgroundPurple)).
		Label("sdk/metric").
		Connect(meter, instruments).
		Connect(instruments, aggregations).
		Connect(aggregations, periodic).
		Connect(aggregations, manual).
		Connect(periodic, slogEmitter)
	sdk.NewGroup("recorder", diagram.WithBackground(diagram.BackgroundGreen)).
		Label("sdk/recorder").
		Connect(recorder, dir)

	contrib := diagram.NewGroup("contrib").Label("contrib").Add(prometheus)

	d.Group(diagram.NewGroup("client").Label("Source").Add(client))
	d.Group(sources)
	d.Group(pipeline)
	d.Group(sdk)
	d.Group(contrib)

	d.Connect(client, file)
	d.Connect(client, reader)
	d.Connect(client, httpSource)
	d.Connect(file, decode)
	d.Connect(reader, decode)
	d.Connect(httpSource, decode)
	d.Connect(exporters, dropQueue)
	d.Connect(exporters, blockQueue)
	d.Connect(exporters, wavExporter)
	d.Connect(dropQueue, meter)
	d.Connect(blockQueue, recorder)
	d.Connect(prometheus, manual)

	return d.Render()
}

func generateEncoding() error {
	d, err := diagram.New(
		diagram.BaseDir("diagrams"),
		diagram.Direction("LR"),
		diagram.Filename("encoding"),
		diagram.Label("Audio API: formats and sample codecs"),
	)
	if err != nil {
		return err
	}

	client := apps.Client.Client().Label("Audio stream or file")
	registry := oci.Governance.Ocid().Label("audio.Decode (magic sniffing)")
	samples := oci.Database.Stream().Label("SampleReader / SampleWriter")
	buffer := oci.Storage.BlockStorage().Label("audio.Buffer")

	wavDecoder := oci.Database.Dis().Label("wav.Decoder")
	wavEncoder := oci.Database.Dis().Label("wav.Encoder")
	riffReader := oci.Governance.Compartments().Label("riff.Reader")
	riffWriter := oci.Governance.Compartments().Label("riff.Writer")

	pcmInt := oci.Database.DatabaseService().Label("Int (8-32 bit, either order)")
	pcmFloat := oci.Database.DatabaseService().Label("Float (32/64 bit)")
	alaw := oci.Database.DatabaseService().Label("A-law")
	mulaw := oci.Database.DatabaseService().Label("mu-law")

	legacy := oci.Governance.Audit().Label("wav.Wav, wav.Stream, data chunks")

	formats := diagram.NewGroup("formats").Label("Formats")
	formats.NewGroup("wav", diagram.WithBackground(diagram.BackgroundPurple)).
		Label("encoding/wav").
		Add(wavDecoder, wavEncoder)
	formats.NewGroup("riff", diagram.WithBackground(diagram.BackgroundYellow)).
		Label("encoding/riff").
		Add(riffReader, riffWriter)

	codecs := diagram.NewGroup("pcm").Label("encoding/pcm").Add(pcmInt, pcmFloat, alaw, mulaw)

	root := diagram.NewGroup("root").Label("audio").Add(registry, samples, buffer)

	old := diagram.NewGroup("legacy", diagram.WithBackground(diagram.BackgroundGreen)).
		Label("Old API (tests only)").
		Add(legacy)

	d.Group(diagram.NewGroup("client").Label("Source").Add(client))
	d.Group(root)
	d.Group(formats)
	d.Group(codecs)
	d.Group(old)

	d.Connect(client, registry)
	d.Connect(registry, wavDecoder)
	d.Connect(wavDecoder, riffReader)
	d.Connect(wavEncoder, riffWriter)
	d.Connect(wavDecoder, pcmInt)
	d.Connect(wavDecoder, pcmFloat)
	d.Connect(wavDecoder, alaw)
	d.Connect(wavDecoder, mulaw)
	d.Connect(wavEncoder, pcmInt)
	d.Connect(wavDecoder, samples)
	d.Connect(samples, wavEncoder)
	d.Connect(samples, buffer)

	return d.Render()
}
