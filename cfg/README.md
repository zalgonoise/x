## cfg

### _a functional options and configuration builder library in Go_

__________

### Concept

`cfg` is a Go library that provides a simple and concise way of adding optional configuration settings to your `Config` 
data structures in Go, leveraging generics and a dash of functional programming patterns.

Generics allow the public functions and types in this library to work with any configuration type, without either 
breaking the static types and without forcing the caller to adhere to a certain type native of the configuration library.

The functional programming patterns are present in the way that the users of the library build their configuration options.
Users provide the logic to change their Config data structure's elements however needed. More on this below.

Furthermore, the strategy is to enable libraries to expose one `Config` type per package, with all private elements; where
consumers have zero control over changing those elements. Instead, the libraries expose configuration functions (`WithXxx`)
that will in turn apply those changes to the `Config` type. Lastly, the `Option` type exported by this library is an 
interface which only contains one method, `apply(T) T`, a private one. This also ensure that consumers of the library cannot
create functions that override elements in the `Config`, given that they are in fact private.

_______

### Motivation

Working and researching the OpenTelemetry Go SDK and API, I noticed this pattern being present in a lot of packages: 
You would find a `Config` type with all private fields, and the constructors are variadic functions that can take any number of
`XxxOption` functions. This `XxxOption` type is also an interface that contains the same private method.

With this it clicked on my head _why are they putting in so much effort in the Options_, thinking that the library could 
just expose `Config` types with public elements, or to set the elements' fields within the actual `XxxOption` function.

Appreciating this approach, I saw it as a great means of allowing configurable, non-mandatory options and settings in a 
constructor (`NewXxx` functions and so on). It makes extra sense in SDKs and APIs that grow in complexity, yet want to 
maintain a controlled level of modularity.

While I was enjoying it through some projects and packages, it became a bit tiresome to create dedicated `Option` 
interfaces in every single package with options. So that is where the experiments with generics began. It worked great.
It works regardless if the `Config` type is or isn't a pointer; it's up to the caller to set it as such in the 
`Option`-returning functions. It saves a lot of effort and has proven useful for something so simple, at least in my
personal opinion.

__________

### Usage

_Content TBA_

### Example

A working example with a pinger application (that checks if GitHub is up or not) is present in the 
[`/examples/pinger` directory](./examples/pinger). In this example, you will find the usual 
[`cmd/pinger`](./examples/pinger/cmd/pinger/main.go) entrypoint where the pinger service is being configured with these
options.

On the other hand, you can also explore [the `ping` package](./examples/pinger/ping) where the service exposes these 
configuration options. The example tries to cover a minimal environment so please take it as only a demonstration.

_Description TBA_

### Disclaimer

_Content TBA_