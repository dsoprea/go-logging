package log

import (
    e "errors"
    "testing"

    "math/rand"
)

// Extends the default environment configuration-provider to set the level to
// "debug" (so none of our messages get filtered).
type testConfigurationProvider struct {
    levelName string
}

func newTestConfigurationProvider(levelName string) ConfigurationProvider {
    if levelName == "" {
        levelName = LevelNameError
    }

    return &testConfigurationProvider{
        levelName: levelName,
    }
}

func (ec *testConfigurationProvider) Format() string {
    return ""
}

func (ec *testConfigurationProvider) DefaultAdapterName() string {
    return ""
}

func (ec *testConfigurationProvider) LevelName() string {
    return ec.levelName
}

func (ec *testConfigurationProvider) IncludeNouns() string {
    return ""
}

func (ec *testConfigurationProvider) ExcludeNouns() string {
    return ""
}

func (ec *testConfigurationProvider) ExcludeBypassLevelName() string {
    return ""
}

/*
type testConfigurationProvider struct {
    EnvironmentConfigurationProvider
}

func (tec *testConfigurationProvider) LevelName() string {
    return LevelNameDebug
}
*/

// A test logging-adapter that sets flags as certain messages are received.
type testLogAdapter struct {
    id int

    debugTriggered   bool
    infoTriggered    bool
    warningTriggered bool
    errorTriggered   bool
}

func newTestLogAdapter() LogAdapter {
    return &testLogAdapter{
        id: rand.Int(),
    }
}

func (tla *testLogAdapter) Debugf(lc *LogContext, message *string) error {
    tla.debugTriggered = true

    return nil
}

func (tla *testLogAdapter) Infof(lc *LogContext, message *string) error {
    tla.infoTriggered = true

    return nil
}

func (tla *testLogAdapter) Warningf(lc *LogContext, message *string) error {
    tla.warningTriggered = true

    return nil
}

func (tla *testLogAdapter) Errorf(lc *LogContext, message *string) error {
    tla.errorTriggered = true

    return nil
}

// Tests

func TestConfigurationOverride(t *testing.T) {
    cs := getConfigState()
    defer func() {
        setConfigState(cs)
    }()

    levelName = "xyz"

    // Overwrite configuration, first thing.
    tcp := newTestConfigurationProvider(LevelNameDebug)
    LoadConfiguration(tcp)

    if levelName != LevelNameDebug {
        t.Error("The test configuration-provider didn't override the level properly.", levelName)
    }
}

func TestConfigurationLevelDirectOverride(t *testing.T) {
    // Overwrite configuration, first thing.
    tcp := newTestConfigurationProvider("")
    LoadConfiguration(tcp)

    ClearAdapters()

    tla1 := newTestLogAdapter()
    AddAdapter("test", tla1)

    l := NewLoggerWithAdapterName("logTest", "test")

    // Usually we don't configure until the first message. Force it.
    l.doConfigure(false)
    tla2 := l.Adapter().(*testLogAdapter)

    if tla2.debugTriggered != false {
        t.Error("Debug flag should've been FALSE initially but wasn't.")
    }

    // Set the level high to prevent logging, first.
    levelName = LevelNameError

    // Force a reconfig (which will bring in the new level).
    l.doConfigure(true)

    // Re-retrieve. This is reconstructed during reconfiguration.
    tla3 := l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")

    if tla3.debugTriggered != false {
        t.Error("Debug message not through but wasn't supposed to.")
    }

    // Now, set the level low to allow logging.
    levelName = LevelNameDebug

    // Force a reconfig (which will bring in the new level).
    l.doConfigure(true)

    // Re-retrieve. This is reconstructed during reconfiguration.
    tla4 := l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")

    if tla4.debugTriggered == false {
        t.Error("Debug message not getting through.")
    }
}

func TestConfigurationLevelProviderOverride(t *testing.T) {
    cs := getConfigState()
    defer func() {
        setConfigState(cs)
    }()

    // Overwrite configuration, first thing.
    tcp := newTestConfigurationProvider("")
    LoadConfiguration(tcp)

    ClearAdapters()

    tla1 := newTestLogAdapter()
    AddAdapter("test", tla1)

    l := NewLoggerWithAdapterName("logTest", "test")

    // Usually we don't configure until the first message. Force it.
    l.doConfigure(false)
    tla2 := l.Adapter().(*testLogAdapter)

    if tla2.debugTriggered != false {
        t.Error("Debug flag should've been FALSE initially but wasn't.")
    }

    // Set the level high to prevent logging, first.
    tcp = newTestConfigurationProvider(LevelNameError)
    LoadConfiguration(tcp)

    // Force a reconfig (which will bring in the new level).
    l.doConfigure(true)

    // Re-retrieve. This is reconstructed during reconfiguration.
    tla3 := l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")

    if tla3.debugTriggered != false {
        t.Error("Debug message not through but wasn't supposed to.")
    }

    // Now, set the level low to allow logging.
    tcp = newTestConfigurationProvider(LevelNameDebug)
    LoadConfiguration(tcp)

    // Force a reconfig (which will bring in the new level).
    l.doConfigure(true)

    // Re-retrieve. This is reconstructed during reconfiguration.
    tla4 := l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")

    if tla4.debugTriggered == false {
        t.Error("Debug message not getting through.")
    }
}

func TestDefaultAdapterAssignment(t *testing.T) {
    SetDefaultAdapterName("")

    ClearAdapters()

    tla := newTestLogAdapter()
    AddAdapter("test1", tla)

    an := GetDefaultAdapterName()
    if an == "" {
        t.Error("Default adapter not set after registration.")
    }

    an = GetDefaultAdapterName()
    if an != "test1" {
        t.Error("Default adapter not set to our adapter after registration.", an)
    }

    SetDefaultAdapterName("test2")
    an = GetDefaultAdapterName()
    if an != "test2" {
        t.Error("SetDefaultAdapterName() did not set default adapter correctly.", an)
    }
}

func TestAdapter(t *testing.T) {
    cs := getConfigState()
    defer func() {
        setConfigState(cs)
    }()

    // Overwrite configuration, first thing.
    tcp := newTestConfigurationProvider(LevelNameDebug)
    LoadConfiguration(tcp)

    ClearAdapters()

    tla1 := newTestLogAdapter()
    AddAdapter("test", tla1)

    l := NewLoggerWithAdapterName("logTest", "test")

    l.doConfigure(false)

    tla2 := l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")
    if tla2.debugTriggered == false {
        t.Error("Debug message not getting through.")
    }

    l.Infof(nil, "Info message")
    if tla2.infoTriggered == false {
        t.Error("Info message not getting through.")
    }

    l.Warningf(nil, "Warning message")
    if tla2.warningTriggered == false {
        t.Error("Warning message not getting through.")
    }

    err := e.New("an error happened")
    l.Errorf(nil, err, "Error message")
    if tla2.errorTriggered == false {
        t.Error("Error message not getting through.")
    }
}

func TestStaticConfiguration(t *testing.T) {
    scp := NewStaticConfigurationProvider()

    cs := getConfigState()
    defer func() {
        setConfigState(cs)
    }()

    scp.SetFormat("aa")
    scp.SetDefaultAdapterName("bb")
    scp.SetLevelName("cc")
    scp.SetIncludeNouns("dd")
    scp.SetExcludeNouns("ee")
    scp.SetExcludeBypassLevelName("ff")

    LoadConfiguration(scp)

    if format != "aa" {
        t.Error("Static configuration provider was not set correctly: format")
    }

    if defaultAdapterName != "bb" {
        t.Error("Static configuration provider was not set correctly: defaultAdapterName")
    }

    if levelName != "cc" {
        t.Error("Static configuration provider was not set correctly: levelName")
    }

    if includeNouns != "dd" {
        t.Error("Static configuration provider was not set correctly: includeNouns")
    }

    if excludeNouns != "ee" {
        t.Error("Static configuration provider was not set correctly: excludeNouns")
    }

    if excludeBypassLevelName != "ff" {
        t.Error("Static configuration provider was not set correctly: excludeBypassLevelName")
    }
}

func TestNoAdapter(t *testing.T) {
    ClearAdapters()

    l := NewLogger("logTest")

    if l.Adapter() != nil {
        t.Error("Logger has an adapter at init when no adapters were available.")
    }

    l.doConfigure(false)

    if l.Adapter() != nil {
        t.Error("Logger has an adapter after configuration no adapters were available.")
    }

    // Should execute, but nothing will happen.
    err := e.New("an error happened")
    l.Errorf(nil, err, "Error message")
}

func TestNewLogger(t *testing.T) {
    noun := "logTest"

    l := NewLogger(noun)
    if l.noun != noun {
        t.Fatalf("Noun not correct: [%s]", l.noun)
    }
}

func TestNewLoggerWithAdapterName(t *testing.T) {
    noun := "logTest"

    originalDefaultAdapterName := GetDefaultAdapterName()

    adapterName := "abcdef"

    cla := NewConsoleLogAdapter()
    AddAdapter(adapterName, cla)

    SetDefaultAdapterName(adapterName)

    defer func() {
        SetDefaultAdapterName(originalDefaultAdapterName)
        delete(adapters, adapterName)
    }()

    l := NewLoggerWithAdapterName(noun, adapterName)
    if l.noun != noun {
        t.Fatalf("Noun not correct: [%s]", l.noun)
    } else if l.an != adapterName {
        t.Fatalf("Adapter-name not correct: [%s]", l.an)
    }
}

func TestIs__unwrapped__hit(t *testing.T) {
    e1 := e.New("test error")
    if Is(e1, e1) != true {
        t.Fatalf("Is() should be true for an unwrapped success")
    }
}

func TestIs__unwrapped__miss(t *testing.T) {
    e1 := e.New("test error")
    e2 := e.New("test error 2")

    if Is(e1, e2) != false {
        t.Fatalf("Is() should be false for an unwrapped failure")
    }
}

func TestIs__wrapped__hit(t *testing.T) {
    e2 := e.New("test error")
    e1 := Wrap(e2)
    if Is(e1, e2) != true {
        t.Fatalf("Is() should be true for a wrapped success")
    }
}

func TestIs__wrapped__miss(t *testing.T) {
    e1 := Errorf("test error")
    e2 := e.New("test error 2")

    if Is(e1, e2) != false {
        t.Fatalf("Is() should be false for a wrapped failure")
    }
}
