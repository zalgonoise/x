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


> `cfg` is served as a Go library. You need to import it in your project.

You can get the library as a Go module, by importing it and running `go mod tidy` on the top-level directory of your
module, or where `go.mod` is placed.

Then, you're able to initialize it with your configuration functions and constructors, when registering option functions
and when applying those options to a new or pre-made configuration data structure.

#### Creating a new config

If a package only contains a type that should be configured, then there is probably room to name your configuration data
structure `Config`. Otherwise, it can be used as the suffix to the target type, like `XxxConfig`.

This type is usually exported (as a public type), but it contains all-private elements. Below is a simple example with a
type holding two optional configuration settings:

```go
package ping

import (
	"time"
)

type Config struct {
	url     string
	timeout time.Duration
}
```

That's it! This `Config` type is the target to your option functions and also the type returned from applying those
same options.

#### Creating option functions

On each available option function you can define whichever type of behavior you desire. It is also an OK place to either
add basic checks (like for zero values) where a [`NoOp` type](./option_noop.go#L5) can be returned. Keep in mind that 
it is not the place to place all your validation where you may wish to return an error. For that it is best to inspect 
the resulting config with validation functions (like `func(c Config) error`).

Below you can see two examples, one library consumer just returns a function of type [`OptionFunc`](./option.go#L10), and the second where 
they use the equivalent [`Register`](./option.go#L19) function instead:

```go
package ping

import (
	"time"
	
	"github.com/zalgonoise/x/cfg"
)

type Config struct {
	url     string
	timeout time.Duration
}

func WithURL(url string) cfg.Option[Config] {
	if url == "" {
		return cfg.NoOp[Config]{}
	}

	// register an option by declaring the returned function as a OptionFunc type
	return cfg.OptionFunc[Config](func(config Config) Config {
		config.url = url

		return config
	})
}

func WithTimeout(dur time.Duration) cfg.Option[Config] {
	if dur <= 0 {
		return cfg.NoOp[Config]{}
	}

	// register an option via the cfg.Register function
	return cfg.Register(func(config Config) Config {
		config.timeout = dur

		return config
	})
}
```

The meat of the function is basically a builder pattern: _I will take your configuration, make some changes to it, and 
return a modified version of it_. While this is also achievable with pointers, Go allows you to do it with non-pointer 
types just as well, provided that you're OK with copying the data structure with each option func. Usually this not
something that is done constantly, thus it shouldn't generally be a problem.

#### Creating a configuration from options

For this action, the library exposes two functions:
- `New[T any](options ...Option[T]) T`: Creates a new configuration data structure from scratch and applies the input 
options on top of it. The resulting type is created with a `*new(T)` call.
- `Set[T any](config T, options ...Option[T]) T`: Applies the input options on top of the input config. This call is 
useful when starting with configuration defaults, to have otherwise unset values with safe defaults.

You're free to use whichever you see fit. Below is an example that takes the `Config` data structure from above and 
applies the input options onto it, while providing safe defaults to begin with:

`ping_config.go`:

```go
package ping

import (
	"time"
	
	"github.com/zalgonoise/x/cfg"
)

const (
	defaultTimeout = 15 * time.Second
	defaultURL     = "https://github.com/"
)

var (
	defaultConfig = Config{
		url:     defaultURL,
		timeout: defaultTimeout,
	}
)

type Config struct {
	url     string
	timeout time.Duration
}

// (...) rest of the configuration logic (option functions)
```

`ping.go`:

```go
package ping

import (
	"time"

	"github.com/zalgonoise/x/cfg"
)

type Checker struct {
	url     string
	timeout time.Duration
}

func NewChecker(options ...cfg.Option[Config]) (*Checker, error) {
	// apply the input options on top of the defined default; the config is a value, not a pointer, in this case.
	config := cfg.Set(defaultConfig, options...)

	if err := validateURL(config); err != nil {
		return nil, err
	}

	// either use the config or pass it along to the data structure if it makes sense that way.
	return &Checker{
		url:     config.url,
		timeout: config.timeout,
	}, nil
}
```

The constructor to the type (`NewChecker` function) spawns the config with a `cfg.Set` call, using the default 
configuration as a base. Note also how validation is done separately -- targeting the resulting config -- which will 
check if the provided URL is OK. That function returns an error which can be useful. Validation is a different topic 
with dedicated logic / workflow if necessary.


#### Using the options

As a caller, you're effortlessly using these constructors and letting your LSP tell you what you can use to configure 
that type. These would be exported functions in the same package as the type and its constructor, that could be nicely 
prefixed (`WithXxx`) to help with alphabetical indexing of the exported types, as the text editors with LSP support and 
IDEs provide context on what they can use.

Below is an example for the same `ping` package referenced above; where in the `main.go` function the caller chooses 
what they want to configure the type with:

```go
package main

import 	"github.com/zalgonoise/x/cfg/examples/pinger/ping"

func main() {
	myURL := "https://github.com/"

	c, err := ping.NewChecker(
		ping.WithURL(myURL),
		// in this case the service has a default for the timeout, but we could
		// override that value if WithTimeout below was not commented out.
		//
		// ping.WithTimeout(30 * time.Second),
	)
	if err != nil {
		// handle error 
	}
	
	// continue to use the checker
}
```

#### Structure and observations

A few important notes that may be useful when creating options of this nature:
- keeping everything in the same package whenever possible; consumers shouldn't have to include multiple imports to 
reach the configuration _and_ the actual type they will be working on.
- Public types, private elements; which ensures that your package is in control of how you can modify the configuration.
- Prefixing option functions with the keyword `With`, if possible; [covered above](#creating-option-functions), but in a 
nutshell enables LSP to list the available options easily.
- Constructors should require mandatory elements as function arguments / parameters, and only the optional configuration
settings served as an `Option[T]` type, where the `Option[T]` slice is variadic (allowing zero elements).
- Validation of the input is performed _after_ the configuration is created from the input parameters and options.


### Example

A working example with a pinger application (that checks if GitHub is up or not) is present in the 
[`/examples/pinger` directory](./examples/pinger). In this example, you will find the usual 
[`cmd/pinger`](./examples/pinger/cmd/pinger/main.go) entrypoint where the pinger service is being configured with these
options. This is the same logic as covered in the [_Using the options_](#using-the-options) section of this document.

On the other hand, you can also explore [the `ping` package](./examples/pinger/ping) where the service exposes these 
configuration options. The example tries to cover a minimal environment so please take it as only a demonstration. This 
is the same logic that serves as example in the first chapters of the [_Usage_](#usage) section.  


### Disclaimer

This is not a one-size-fits-all solution! Please take your time to evaluate it for your own needs with due diligence. 
While having _a library for this and a library for that_ is pretty nice, it could potentially be only overhead hindering
the true potential of your app! Be sure to read the code that you are using to be a better judge if it is a good fit for 
your project. With that in mind, I hope you enjoy this library. Feel free to contribute by filing either an issue or a 
pull request.