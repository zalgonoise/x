## Benchmarking `zalgonoise/x/audio/wav`

____________

### Overview

This document will display benchmark results for encoding and decoding WAV files using `zalgonoise/x/audio/wav` and 
`go-audio/wav`, for performance comparison purposes.

The benchmark tests are based on several, differently encoded WAV (short) files using both libraries. While the testdata
is exactly the same, the flow of encoding and decoding the audio from and to bytes and a PCM buffer is very different in
both libraries; where my (`zalgonoise/x/audio/wav`) implementation tries to be as minimal as possible.

Regarding decoding (bytes to PCM buffer), there is a clear improvement with a range of 500~3000 times less allocated 
bytes; and between half and a third fewer allocations per operation.

On encoding, there is a shorter gap when it comes to bytes allocated, between 150 and 250 times fewer; but up to 
4000 times fewer allocations per operation.

The strong points about `zalgonoise/x/audio/wav` are the decoding performance, especially on log(2) bit rates 
(8, 16 and 32); as they are naturally converted to the appropriate integer type in Go. The weak points are the 24-bit 
conversions (encoding or decoding), in which the bit shifting slows the processing down in comparison -- yet still (way)
faster than `go-audio/wav`.

__________________

### 2023-03-10 Results

```
❯ go test -bench . -benchtime=5s -benchmem  . | prettybench 
goos: linux
goarch: amd64
pkg: github.com/zalgonoise/x/benchmark/audio
cpu: AMD Ryzen 3 PRO 3300U w/ Radeon Vega Mobile Gfx
PASS
benchmark                                           iter           time/iter     bytes alloc            allocs
---------                                           ----           ---------     -----------            ------
BenchmarkGoAudioWav/DecodeMono8Bit44100Hz-4        23246     287551.00 ns/op     230153 B/op      38 allocs/op
BenchmarkGoAudioWav/EncodeMono8Bit44100Hz-4         3285    1767939.00 ns/op    2663980 B/op    9376 allocs/op
BenchmarkGoAudioWav/DecodeMono16Bit44100Hz-4       15481     385230.00 ns/op     230109 B/op      38 allocs/op
BenchmarkGoAudioWav/EncodeMono16Bit44100Hz-4        2059    2668615.00 ns/op    5337248 B/op   18328 allocs/op
BenchmarkGoAudioWav/DecodeMono24Bit44100Hz-4       17887     336327.00 ns/op     230147 B/op      38 allocs/op
BenchmarkGoAudioWav/EncodeMono24Bit44100Hz-4        1467    3975619.00 ns/op    8230847 B/op   28072 allocs/op
BenchmarkGoAudioWav/DecodeMono32Bit44100Hz-4       18397     326751.00 ns/op     230111 B/op      38 allocs/op
BenchmarkGoAudioWav/EncodeMono32Bit44100Hz-4        1670    3466025.00 ns/op   10667468 B/op   18724 allocs/op
BenchmarkGoAudioWav/DecodeMono32Bit96000Hz-4        7798     690854.00 ns/op     582817 B/op      40 allocs/op
BenchmarkGoAudioWav/EncodeMono32Bit96000Hz-4         938    6978697.00 ns/op   23207326 B/op   40728 allocs/op
BenchmarkGoAudioWav/DecodeMono32Bit192000Hz-4       3808    1584821.00 ns/op    1558901 B/op      44 allocs/op
BenchmarkGoAudioWav/EncodeMono32Bit192000Hz-4        457   13003924.00 ns/op   46414222 B/op   81428 allocs/op
BenchmarkGoAudioWav/DecodeMono8Bit176400Hz-4        3901    1544544.00 ns/op    1559057 B/op      44 allocs/op
BenchmarkGoAudioWav/EncodeMono8Bit176400Hz-4         798    7366921.00 ns/op   10630072 B/op   37422 allocs/op
BenchmarkGoAudioWav/DecodeStereo8Bit44100Hz-4       8242     686695.00 ns/op     582996 B/op      40 allocs/op
BenchmarkGoAudioWav/EncodeStereo8Bit44100Hz-4       1705    3521123.00 ns/op    5319344 B/op   18724 allocs/op
BenchmarkGoAudioWav/DecodeStereo16Bit44100Hz-4      7130     785914.00 ns/op     583297 B/op      40 allocs/op
BenchmarkGoAudioWav/EncodeStereo16Bit44100Hz-4      1226    4732128.00 ns/op   10665885 B/op   36628 allocs/op
BenchmarkGoAudioWav/DecodeStereo24Bit44100Hz-4      8509     668005.00 ns/op     583129 B/op      40 allocs/op
BenchmarkGoAudioWav/EncodeStereo24Bit44100Hz-4       738    7980315.00 ns/op   16453076 B/op   56116 allocs/op
BenchmarkGoAudioWav/DecodeStereo32Bit44100Hz-4      8695     663942.00 ns/op     583098 B/op      40 allocs/op
BenchmarkGoAudioWav/EncodeStereo32Bit44100Hz-4       927    6284064.00 ns/op   21326318 B/op   37420 allocs/op
BenchmarkXAudioWav/DecodeMono8Bit44100Hz-4       1692238       3508.00 ns/op        464 B/op      14 allocs/op
BenchmarkXAudioWav/EncodeMono8Bit44100Hz-4        936888       5950.00 ns/op       9872 B/op      11 allocs/op
BenchmarkXAudioWav/DecodeMono16Bit44100Hz-4      1659931       3444.00 ns/op        464 B/op      14 allocs/op
BenchmarkXAudioWav/EncodeMono16Bit44100Hz-4        54085     112225.00 ns/op      38544 B/op      12 allocs/op
BenchmarkXAudioWav/DecodeMono24Bit44100Hz-4        60006     100706.00 ns/op      41424 B/op      15 allocs/op
BenchmarkXAudioWav/EncodeMono24Bit44100Hz-4        60092     106771.00 ns/op      57744 B/op      12 allocs/op
BenchmarkXAudioWav/DecodeMono32Bit44100Hz-4      1677364       3589.00 ns/op        464 B/op      14 allocs/op
BenchmarkXAudioWav/EncodeMono32Bit44100Hz-4        60006     100354.00 ns/op      82320 B/op      12 allocs/op
BenchmarkXAudioWav/DecodeMono32Bit96000Hz-4      1693435       3544.00 ns/op        464 B/op      14 allocs/op
BenchmarkXAudioWav/EncodeMono32Bit96000Hz-4        29467     204116.00 ns/op     164240 B/op      12 allocs/op
BenchmarkXAudioWav/DecodeMono32Bit192000Hz-4     1689103       3488.00 ns/op        464 B/op      14 allocs/op
BenchmarkXAudioWav/EncodeMono32Bit192000Hz-4       14797     401933.00 ns/op     328080 B/op      12 allocs/op
BenchmarkXAudioWav/DecodeMono8Bit176400Hz-4      1824980       3523.00 ns/op        464 B/op      14 allocs/op
BenchmarkXAudioWav/EncodeMono8Bit176400Hz-4       376702      16023.00 ns/op      41360 B/op      11 allocs/op
BenchmarkXAudioWav/DecodeStereo8Bit44100Hz-4     1648074       3539.00 ns/op        464 B/op      14 allocs/op
BenchmarkXAudioWav/EncodeStereo8Bit44100Hz-4      686534       8754.00 ns/op      19472 B/op      11 allocs/op
BenchmarkXAudioWav/DecodeStereo16Bit44100Hz-4    1659889       3561.00 ns/op        464 B/op      14 allocs/op
BenchmarkXAudioWav/EncodeStereo16Bit44100Hz-4      31879     187989.00 ns/op      82320 B/op      12 allocs/op
BenchmarkXAudioWav/DecodeStereo24Bit44100Hz-4      31041     188902.00 ns/op      82384 B/op      15 allocs/op
BenchmarkXAudioWav/EncodeStereo24Bit44100Hz-4      26809     214831.00 ns/op     115088 B/op      12 allocs/op
BenchmarkXAudioWav/DecodeStereo32Bit44100Hz-4    1810458       3521.00 ns/op        464 B/op      14 allocs/op
BenchmarkXAudioWav/EncodeStereo32Bit44100Hz-4      31279     192220.00 ns/op     164240 B/op      12 allocs/op
ok      github.com/zalgonoise/x/benchmark/audio 338.122s
```