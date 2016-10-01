package log

import (
    "text/template"

    e "errors"
    "fmt"
    "bytes"
    "strings"

    "golang.org/x/net/context"

    "github.com/go-errors/errors"
)

// Config severity integers.
const (
    LevelDebug = iota
    LevelInfo = iota
    LevelWarning = iota
    LevelError = iota
)

// Config severity names.
const (
    LevelNameDebug = "debug"
    LevelNameInfo = "info"
    LevelNameWarning = "warning"
    LevelNameError = "error"
)

// Seveirty name->integer map.
var (
    LevelNameMap = map[string]int {
        LevelNameDebug: LevelDebug,
        LevelNameInfo: LevelInfo,
        LevelNameWarning: LevelWarning,
        LevelNameError: LevelError,
    }
)

// Errors
var (
    ErrAdapterMakerAlreadyDefined = e.New("adapter-maker already defined")
    ErrFormatEmpty = e.New("format is empty")
    ErrExcludeLevelNameInvalid = e.New("exclude bypass-level is invalid")
    ErrNoAdapterConfigured = e.New("no default adapter configured")
    ErrAdapterMakerIsNil = e.New("adapter-maker is nil")
    ErrConfigurationNotLoaded = e.New("can not configure because configuration is not loaded")
)

// Other
var (
    includeFilters = make(map[string]bool)
    useIncludeFilters = false
    excludeFilters = make(map[string]bool)
    useExcludeFilters = false

    makers = make(map[string]AdapterMaker)

// TODO(dustin): !! Finish implementing this.
    excludeBypassLevel = -1
)

// Add global include filter.
func AddIncludeFilter(noun string) {
    includeFilters[noun] = true
    useIncludeFilters = true
}

// Remove global include filter.
func RemoveIncludeFilter(noun string) {
    delete(includeFilters, noun)
    if len(includeFilters) == 0 {
        useIncludeFilters = false
    }
}

// Add global exclude filter.
func AddExcludeFilter(noun string) {
    excludeFilters[noun] = true
    useExcludeFilters = true
}

// Remove global exclude filter.
func RemoveExcludeFilter(noun string) {
    delete(excludeFilters, noun)
    if len(excludeFilters) == 0 {
        useExcludeFilters = false
    }
}

type AdapterMaker interface {
    New() LogAdapter
}

func AddAdapterMaker(name string, am AdapterMaker) {
    if _, found := makers[name]; found == true {
        Panic(ErrAdapterMakerAlreadyDefined)
    }

    if am == nil {
        Panic(ErrAdapterMakerIsNil)
    }

    makers[name] = am

    if adapterName == "" {
        adapterName = name
    }
}

func ClearAdapters() {
    makers = make(map[string]AdapterMaker)
    adapterName = ""
}

type LogAdapter interface {
    Debugf(lc *LogContext, message *string) error
    Infof(lc *LogContext, message *string) error
    Warningf(lc *LogContext, message *string) error
    Errorf(lc *LogContext, message *string) error
}

// TODO(dustin): !! Also populate whether we've bypassed an exception so that 
//                  we can add a template macro to prefix an exclamation of 
//                  some sort.
type MessageContext struct {
    Noun *string
    Message *string
    ExcludeBypass bool
}

type LogContext struct {
    Logger *Logger
    Ctx context.Context
}

type Logger struct {
    isConfigured bool
    an string
    la LogAdapter
    t *template.Template
    systemLevel int
    noun string
}

// This is very basic. It might be called at the module level at a point where 
// configuration still hasn't been equipped or adapters registered. Those are 
// done lazily (see `doConfigure`).
func NewLoggerWithAdapter(noun string, adapterName string) *Logger {
    l := &Logger{
        noun: noun,
        an: adapterName,

        // We set this lazily since this function can, and will likely, be 
        // called at the module-level and we won't have any makers registered 
        // yet.
        la: nil,
    }

    return l
}

// TODO(dustin): !! We need to cement the plan for how to configure the core project (currently use os.GetEnv() because that's what AppEngine requires). Then, we need to adopt it and impress the important of setting the default adapter-name there if the application will be depending on it. Otherwise, there's no way for us to have this when NewLogger() is called.
// TODO(dustin): !! We might consider using a interface-driven design for providing the configuration. We can only do this if the NewLogger call can avoid using configuration (since that will be called a bunch of times before we have the opportunity to read the configuration.
func NewLogger(noun string) *Logger {
    if adapterName == "" {
        Panic(ErrNoAdapterConfigured)
    }

    return NewLoggerWithAdapter(noun, adapterName)
}

func (l *Logger) Noun() string {
    return l.noun
}

func (l *Logger) Adapter() LogAdapter {
    return l.la
}

func (l *Logger) doConfigure(force bool) {
    if l.isConfigured == false || force == true {
        if IsConfigurationLoaded() == false {
            Panic(ErrConfigurationNotLoaded)
        }

        am, found := makers[l.an]
        if found == false {
            Panic(fmt.Errorf("adapter is not valid: %s", l.an))
        }

        l.la = am.New()

        // Set the level.

        systemLevel, found := LevelNameMap[levelName]
        if found == false {
            panic(fmt.Errorf("log-level not valid: [%s]", levelName))
        }

        l.systemLevel = systemLevel

        // Set the form.

        if format == "" {
            panic(ErrFormatEmpty)
        }

        l.SetFormat(format)

        l.isConfigured = true
    }
}

func (l *Logger) SetFormat(format string) {
    if t, err := template.New("logItem").Parse(format); err != nil {
        panic(err)
    } else {
        l.t = t
    }
}

func (l *Logger) flattenMessage(lc *MessageContext, format *string, args []interface{}) (string, error) {
    m := fmt.Sprintf(*format, args...)

    lc.Message = &m
    
    var b bytes.Buffer
    if err := l.t.Execute(&b, *lc); err != nil {
        return "", err
    }

    return b.String(), nil
}

func (l *Logger) allowMessage(noun string, level int) bool {
    if _, found := includeFilters[noun]; found == true {
        return true
    }

    // If we didn't hit an include filter and we *had* include filters, filter 
    // it out.
    if useIncludeFilters == true {
        return false
    }

    if _, found := excludeFilters[noun]; found == true {
        return false
    }

    return true
}

func (l *Logger) makeLogContext(ctx context.Context) *LogContext {
    return &LogContext{
        Ctx: ctx,
        Logger: l,
    }
}

type LogMethod func(lc *LogContext, message *string) error

func (l *Logger) log(ctx context.Context, level int, lm LogMethod, format string, args []interface{}) error {
    if l.systemLevel > level {
        return nil
    }

    // Preempt the normal filter checks if we can unconditionally allow at a 
    // certain level and we've hit that level.
    //
    // Notice that this is only relevant if the system-log level is letting 
    // *anything* show logs at the level we came in with.
    canExcludeBypass := level >= excludeBypassLevel && excludeBypassLevel != -1
    didExcludeBypass := false

    n := l.Noun()

    if(l.allowMessage(n, level) == false) {
        if canExcludeBypass == false {
            return nil
        } else {
            didExcludeBypass = true
        }
    }

    lc := &MessageContext{
        Noun: &n,
        ExcludeBypass: didExcludeBypass,
    }

    if s, err := l.flattenMessage(lc, &format, args); err != nil {
        return err
    } else {
        lc := l.makeLogContext(ctx)
        if err := lm(lc, &s); err != nil {
            panic(err)
        }

        return e.New(s)
    }
}

func (l *Logger) Debugf(ctx context.Context, format string, args ...interface{}) {
    l.doConfigure(false)
    l.log(ctx, LevelDebug, l.la.Debugf, format, args)
}

func (l *Logger) Infof(ctx context.Context, format string, args ...interface{}) {
    l.doConfigure(false)
    l.log(ctx, LevelInfo, l.la.Infof, format, args)
}

func (l *Logger) Warningf(ctx context.Context, format string, args ...interface{}) {
    l.doConfigure(false)
    l.log(ctx, LevelWarning, l.la.Warningf, format, args)
}

func (l *Logger) mergeStack(err interface{}, format string, args []interface{}) (string, []interface{}) {
    if format != "" {
        format += "\n%s"
    } else {
        format = "%s"
    }

    var stackified *errors.Error
    stackified, ok := err.(*errors.Error)
    if ok == false {
        stackified = errors.Wrap(err, 2)
    }

    args = append(args, stackified.ErrorStack())

    return format, args
}

func (l *Logger) Errorf(ctx context.Context, err interface{}, format string, args ...interface{}) {
    l.doConfigure(false)

    format, args = l.mergeStack(err, format, args)
    l.log(ctx, LevelError, l.la.Errorf, format, args)
}

func (l *Logger) ErrorIff(ctx context.Context, err interface{}, format string, args ...interface{}) {
    if err == nil {
        return
    }

    l.Errorf(ctx, err, format, args...)
}

func (l *Logger) Panicf(ctx context.Context, err interface{}, format string, args ...interface{}) {
    l.doConfigure(false)

    format, args = l.mergeStack(err, format, args)
    errFlat := l.log(ctx, LevelError, l.la.Errorf, format, args)
    panic(errFlat)
}

func (l *Logger) PanicIff(ctx context.Context, err interface{}, format string, args ...interface{}) {
    if err == nil {
        return
    }

    _, ok := err.(*errors.Error)
    if ok == true {
        panic(err)
    } else {
        panic(errors.Wrap(err, 1))
    }
}

func Panic(err interface{}) {
    _, ok := err.(*errors.Error)
    if ok == true {
        panic(err)
    } else {
        panic(errors.Wrap(err, 1))
    }
}

func Wrap(err interface{}) *errors.Error {
    es, ok := err.(*errors.Error)
    if ok == true {
        return es
    } else {
        return errors.Wrap(err, 1)
    }
}

func PanicIf(err interface{}) {
    if err == nil {
        return
    }

    _, ok := err.(*errors.Error)
    if ok == true {
        panic(err)
    } else {
        panic(errors.Wrap(err, 1))
    }
}

func init() {
    if format == "" {
        format = defaultFormat
    }

    if levelName == "" {
        levelName = defaultLevelName
    }

    if includeNouns != "" {
        for _, noun := range strings.Split(includeNouns, ",") {
            AddIncludeFilter(noun)
        }
    }

    if excludeNouns != "" {
        for _, noun := range strings.Split(excludeNouns, ",") {
            AddExcludeFilter(noun)
        }
    }

    if excludeBypassLevelName != "" {
        var found bool
        if excludeBypassLevel, found = LevelNameMap[excludeBypassLevelName]; found == false {
            panic(ErrExcludeLevelNameInvalid)
        }
    }
}
