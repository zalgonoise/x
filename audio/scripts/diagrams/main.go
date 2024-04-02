package main

import (
	"context"
	"log/slog"

	"github.com/zalgonoise/go-diagrams/diagram"
	"github.com/zalgonoise/go-diagrams/nodes/apps"
	"github.com/zalgonoise/go-diagrams/nodes/oci"
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
		diagram.Label("Audio SDK components and workflow"),
	)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func generateEncoding() error {
	d, err := diagram.New(
		diagram.BaseDir("diagrams"),
		diagram.Direction("LR"),
		diagram.Filename("encoding"),
		diagram.Label("Audio API - encoding"),
	)
	if err != nil {
		return err
	}

	wavHeader := oci.Database.Stream().Label("Header")
	wavChunks := oci.Database.Stream().Label("Data Chunks")

	wavEnc := diagram.NewGroup("wav").Label("WAV").
		Add(wavHeader, wavChunks)

	chunkHeader := oci.Database.Stream().Label("Header")
	chunkData := oci.Database.Stream().Label("Data")
	chunkConverter := oci.Database.Stream().Label("Converter")

	wavChunk := diagram.NewGroup("chunk").Label("Data Chunk").
		Add(chunkHeader, chunkConverter, chunkData)

	junkHeader := oci.Database.Stream().Label("Header")
	junkData := oci.Database.Stream().Label("Data")

	wavJunk := diagram.NewGroup("junk").Label("Junk Chunk").
		Add(junkHeader, junkData)

	ringHeader := oci.Database.Stream().Label("Header")
	ringData := oci.Database.Stream().Label("Data")
	ringConverter := oci.Database.Stream().Label("Converter")

	wavRingChunk := diagram.NewGroup("ring_chunk").Label("Ring-buffer Data Chunk").
		Add(ringHeader, ringConverter, ringData)

	conv8bit := oci.Database.Stream().Label("8bit PCM")
	conv16bit := oci.Database.Stream().Label("16bit PCM")
	conv24bit := oci.Database.Stream().Label("32bit PCM")
	conv32bit := oci.Database.Stream().Label("64bit PCM")
	conv32bitFloat := oci.Database.Stream().Label("32bit FPA")
	conv64bitFloat := oci.Database.Stream().Label("64bit FPA")

	convPlaceholder := oci.Database.Stream().Label("Raw Audio Data Converter")
	converters := diagram.NewGroup("conv").Label("Encodings").
		Add(conv8bit, conv16bit, conv24bit, conv32bit, conv32bitFloat, conv64bitFloat)
	converters.ConnectAllFrom(convPlaceholder.ID())

	converterGroup := diagram.NewGroup("conv_group").Label("Audio Converters").
		Add(convPlaceholder)
	converterGroup.Group(converters)

	wavGroup := diagram.NewGroup("wav_group").Label("WAV encoding")

	wavIO := oci.Database.Stream().Label("WAV I/O")
	wavStreamIO := oci.Database.Stream().Label("WAV Stream I/O")
	wavIOGroup := diagram.NewGroup("wav_io").Label("I/O").Add(wavIO, wavStreamIO)

	wavGroup.Group(wavEnc)
	wavGroup.Group(wavIOGroup)
	wavGroup.Connect(wavIO, wavChunks)
	wavGroup.Connect(wavStreamIO, wavChunks)
	wavGroup.Group(wavChunk)
	wavGroup.Group(wavRingChunk)
	wavGroup.Group(wavJunk)
	wavGroup.Group(converterGroup)
	wavGroup.Connect(chunkConverter, convPlaceholder)
	wavGroup.Connect(ringConverter, convPlaceholder)
	wavGroup.Connect(wavChunks, chunkData)
	wavGroup.Connect(wavChunks, ringData)
	wavGroup.Connect(wavChunks, junkData)

	client := apps.Client.Client().Label("Audio source")
	clientGroup := diagram.NewGroup("client").Label("Source")

	audioFile := oci.Database.Stream().Label("Audio file or buffer")
	audioStream := oci.Database.Stream().Label("Audio stream")

	d.Group(wavGroup)
	d.Group(clientGroup)
	d.Connect(client, audioFile)
	d.Connect(client, audioStream)
	d.Connect(audioFile, wavIO)
	d.Connect(audioStream, wavStreamIO)

	if err := d.Render(); err != nil {
		return err
	}

	return nil
}
