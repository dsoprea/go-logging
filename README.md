|Build\_Status|

## Introduction

Go's logging is pretty basic by design. Go under AppEngine is even more stripped down. For example, there is no logging struct (`Logger` under traditional Go) and there is no support for prefixes. This package not only equips an AppEngine application with both, but adds include/exclude filters, pluggable logging adapters, and configuration-driven design. This allows you to be able to turn on and off logging for certain packages/files or dependencies as well as to be able to do it from configuration (the YAML files, using `os.Getenv()`).


## Getting Started

The simplest, possible example:

```go
package thispackage

import (
    "golang.org/x/net/context"

    "github.com/dsoprea/go-logging"
    "github.com/dsoprea/go-appengine-logging"
)

var (
    thisfile_log = log.NewLogger("thisfile")
)

func a_cry_for_help(ctx context.Context) {
    thisfile_log.Errorf(ctx, "How big is my problem: %s", "pretty big")
}

func init() {
    log.AddAdapterMaker("appengine", aelog.NewAppengineAdapterMaker())
}
```

Notice two things:

1. We configure the "appengine" adapter factory. The first adapter registered will be used be default.
2. We pass in-the name of a prefix (what we refer to as a "noun") to `log.NewLogger()`. This is a simple, descriptive name that represents the current body of logic. We recommend that you define a different log for every file at the package level, but it is your choice if you want to go with this methodology, share the same logger over the entire package, define one for each struct, etc..

### Example Output

Example output from a real application (not from the above):

```
2016/09/09 12:57:44 DEBUG: user: User revisiting: [test@example.com]
2016/09/09 12:57:44 DEBUG: context: Session already inited: [DCRBDGRY6RMWANCSJXVLD7GULDH4NZEB6SBAQ3KSFIGA2LP45IIQ]
2016/09/09 12:57:44 DEBUG: session_data: Session save not necessary: [DCRBDGRY6RMWANCSJXVLD7GULDH4NZEB6SBAQ3KSFIGA2LP45IIQ]
2016/09/09 12:57:44 DEBUG: context: Got session: [DCRBDGRY6RMWANCSJXVLD7GULDH4NZEB6SBAQ3KSFIGA2LP45IIQ]
2016/09/09 12:57:44 DEBUG: session_data: Found user in session.
2016/09/09 12:57:44 DEBUG: cache: Cache miss: [geo.geocode.reverse:dhxp15x]
```


## Backwards Compatibility

This project is the result of a split that occurred early in the life of [go-appengine-logging](https://github.com/dsoprea/go-appengine-logging). That was the original project and it had a limited amount of functionality in the beginning. Once it got implemented and experienced some natural growth it became very, very useful and, simultaneously, non-trivial enough that we weren't thrilled about duplicating it into a non-AppEngine-specific project (the original plan).

Without getting into too much unnecessary detail, the following decisions were made:

1. The main project would be *go-logging*. **Any existing references to *go-appengine-logging* in any projects that depend on this would have to be updated.**
2. Because the *go-logging* code could no longer be intrinsically aware of *go-appengine-logging*, **applications that use *go-logging* must specifically import *go-appengine-logging* and register the one with the other**.

This will break anyone that is not vendoring *go-appengine-logging*, but, as that was an AppEngine-specific project and AE projects are predisposed to vendoring everything, this won't be an issue unless you update. Plus, that project is still pretty recent and adoption is still jsut ramping up.

Sorry for any inconvenience. The original project completely changed the debugging experience for AppEngine and I wanted to bring that over to general Go development.


## Adapters

This project provides one built-in logging adapter: "console", which prints to the screen. To register it:

```go
cam := log.NewConsoleAdapterMaker()
log.AddAdapterMaker("console", cam)
```

If you would like to implement your own logger, just create a struct type that satisfies the LogAdapter interface.

```go
type LogAdapter interface {
    Debugf(lc *LogContext, message *string) error
    Infof(lc *LogContext, message *string) error
    Warningf(lc *LogContext, message *string) error
    Errorf(lc *LogContext, message *string) error
}
```

The *LogContext* struct passed in provides additional information that you may need in order to do what you need to do:

```go
type LogContext struct {
    Logger *Logger
    Ctx context.Context
}
```

Note that *Logger* represents your Logger instance. It exports `Noun()` in the event you want to discriminate where your log entries go. It also exports `Adapter()` for if you need access to the adapter instance.

Adapter example:

```go
type DummyLogAdapter struct {

}

func (dla *DummyLogAdapter) Debugf(lc *LogContext, message *string) error {
    
}

func (dla *DummyLogAdapter) Infof(lc *LogContext, message *string) error {
    
}

func (dla *DummyLogAdapter) Warningf(lc *LogContext, message *string) error {
    
}

func (dla *DummyLogAdapter) Errorf(lc *LogContext, message *string) error {
    
}

func (dla *DummyLogAdapter) Criticalf(lc *LogContext, message *string) error {
    
}
```

There are a couple of ways to tell Logger to use a specific adapter:

1. Instead of calling `log.NewLogger(noun string)`, call `log.NewLoggerWithAdapter(noun string, adapterName string)` and provide a struct of your adapter type.
2. Register a factory type for your adapter and set the name of the adapter into your YAML configuration (under `env_variables`).


The factory must satisfy the *AdapterMaker* interface:

```go
type AdapterMaker interface {
    New() LogAdapter
}
```

An example factory and registration of the factory:

```go
type DummyLogAdapterMaker struct {
    
}

func (dlam *DummyLogAdapterMaker) New() log.LogAdapter {
    return new(DummyLogAdapter)
}
```

We then recommending registering it from the `init()` function of the fiel that defines the maker type:

```go
func init() {
    log.AddAdapterMaker("dummy", new(DummyLogAdapterMaker))
}
```

We discuss how to then reference the adapter-maker from configuration in the "Configuration" section below.


## Filters

We support the ability to exclusively log for a specific set of nouns (we'll exclude any not specified):

```go
log.AddIncludeFilter("nountoshow1")
log.AddIncludeFilter("nountoshow2")
```

Depending on your needs, you might just want to exclude a couple and include the rest:

```go
log.AddExcludeFilter("nountohide1")
log.AddExcludeFilter("nountohide2")
```

We'll first hit the include-filters. If it's in there, we'll forward the log item to the adapter. If not, and there is at least one include filter in the list, we won't do anything. If the list of include filters is empty but the noun appears in the exclude list, we won't do anything.

It is a good convention to exclude the nouns of any library you are writing whose logging you do not want to generally be aware of unless you are debugging. You might call `AddExcludeFilter()` from the `init()` function at the bottom of those files unless there is some configuration variable, such as "(LibraryNameHere)DoShowLogging", that has been defined and set to TRUE.


## Configuration

The following configuration items are available:

- *LogFormat*: The default format used to build the message that gets sent to the adapter. It is assumed that the adapter already prefixes the message with time and log-level (since the default AppEngine logger does). The default value is: `{{.Noun}}:{{if eq .ExcludeBypass true}} [BYPASS]{{end}} {{.Message}}`. The available tokens are "Noun", "ExcludeBypass", and "Message".
- *LogAdapterName*: The name of the adapter to use when NewLogger() is called.
- *LogLevelName*: The priority-level of messages permitted to be logged (all others will be discarded). By default, it is "info". Other levels are: "debug", "warning", "error", "critical"
- *LogIncludeNouns*: Comma-separated list of nouns to log for. All others will be ignored.
- *LogExcludeNouns*: Comma-separated list on nouns to exclude from logging.
- *LogExcludeBypassLevelName*: The log-level at which we will show logging for nouns that have been excluded. Allows you to hide excessive, unimportant logging for nouns but to still see their warnings, errors, etc...

You provide the configuration by setting a configuration-provider. Configuration providers must satisfy the `ConfigurationProvider` interface. The following are provided with the project:

- `EnvironmentConfigurationProvider`: Read values from the environment.
- `StaticConfigurationProvider`: Set values directly on the struct.

Environments such as AppEngine work best with `EnvironmentConfigurationProvider` as this is generally how configuration is exposed *by* AppEngine *to* the application. You can define this configuration directly in *that* configuration.

By default, the environment configuration-provider is created and applied immediately. You may load another prior to logging any messages. A `Logger` instance will apply configuration the first time it is used.

If a configuration-provider does not provide a log-level or format, they will be defaulted (or left alone, if already set). If it does not provide an adapter-name, the adapter-name of the first registered adapter will be used.

Usage instructions of both follow.


### Environment-Based Configuration

```go
ecp := log.NewEnvironmentConfigurationProvider()
log.LoadConfiguration(ecp)
```


### Static Configuration

```go
scp := log.NewStaticConfigurationProvider()
scp.SetLevelName(log.LevelNameWarning)

log.LoadConfiguration(scp)
```
