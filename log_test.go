package log

import (
    "math/rand"

    "testing"
    e "errors"
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

func (ec *testConfigurationProvider) AdapterName() string {
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

    debugTriggered bool
    infoTriggered bool
    warningTriggered bool
    errorTriggered bool
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

// Factory for the test logging-adapter.
type testAdapterMaker struct {

}

func newTestAdapterMaker() *testAdapterMaker {
    return new(testAdapterMaker)
}

func (cam testAdapterMaker) New() LogAdapter {
    return &testLogAdapter{
        id: rand.Int(),
    }
}

// Tests

func TestConfigurationOverride(t *testing.T) {
    oldLevelName := levelName
    
    levelName = "xyz"
    defer func() {
        levelName = oldLevelName
    }()

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

    tam := newTestAdapterMaker()
    AddAdapterMaker("test", tam)

    l := NewLoggerWithAdapter("logTest", "test")

    // Usually we don't configure until the first message. Force it.
    l.doConfigure(false)
    tla := l.Adapter().(*testLogAdapter)

    if tla.debugTriggered != false {
        t.Error("Debug flag should've been FALSE initially but wasn't.")
    }

    // Set the level high to prevent logging, first.
    levelName = LevelNameError
    
    // Force a reconfig (which will bring in the new level).
    l.doConfigure(true)
    
    // Re-retrieve. This is reconstructed during reconfiguration.
    tla = l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")

    if tla.debugTriggered != false {
        t.Error("Debug message not through but wasn't supposed to.")
    }

    // Now, set the level low to allow logging.
    levelName = LevelNameDebug

    // Force a reconfig (which will bring in the new level).
    l.doConfigure(true)

    // Re-retrieve. This is reconstructed during reconfiguration.
    tla = l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")

    if tla.debugTriggered == false {
        t.Error("Debug message not getting through.")
    }
}

func TestConfigurationLevelProviderOverride(t *testing.T) {
    // Overwrite configuration, first thing.
    tcp := newTestConfigurationProvider("")
    LoadConfiguration(tcp)

    ClearAdapters()

    tam := newTestAdapterMaker()
    AddAdapterMaker("test", tam)

    l := NewLoggerWithAdapter("logTest", "test")

    // Usually we don't configure until the first message. Force it.
    l.doConfigure(false)
    tla := l.Adapter().(*testLogAdapter)

    if tla.debugTriggered != false {
        t.Error("Debug flag should've been FALSE initially but wasn't.")
    }

    // Set the level high to prevent logging, first.
    tcp = newTestConfigurationProvider(LevelNameError)
    LoadConfiguration(tcp)
    
    // Force a reconfig (which will bring in the new level).
    l.doConfigure(true)
    
    // Re-retrieve. This is reconstructed during reconfiguration.
    tla = l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")

    if tla.debugTriggered != false {
        t.Error("Debug message not through but wasn't supposed to.")
    }

    // Now, set the level low to allow logging.
    tcp = newTestConfigurationProvider(LevelNameDebug)
    LoadConfiguration(tcp)

    // Force a reconfig (which will bring in the new level).
    l.doConfigure(true)

    // Re-retrieve. This is reconstructed during reconfiguration.
    tla = l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")

    if tla.debugTriggered == false {
        t.Error("Debug message not getting through.")
    }
}

func TestDefaultAdapterAssignment(t *testing.T) {
    SetDefaultAdapterName("")

    ClearAdapters()

    tam := newTestAdapterMaker()
    AddAdapterMaker("test1", tam)

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
    // Overwrite configuration, first thing.
    tcp := newTestConfigurationProvider(LevelNameDebug)
    LoadConfiguration(tcp)

    ClearAdapters()

    tam := newTestAdapterMaker()
    AddAdapterMaker("test", tam)

    l := NewLoggerWithAdapter("logTest", "test")

    l.doConfigure(false)

    tla := l.Adapter().(*testLogAdapter)

    l.Debugf(nil, "Debug message")
    if tla.debugTriggered == false {
        t.Error("Debug message not getting through.")
    }

    l.Infof(nil, "Info message")
    if tla.infoTriggered == false {
        t.Error("Info message not getting through.")
    }

    l.Warningf(nil, "Warning message")
    if tla.warningTriggered == false {
        t.Error("Warning message not getting through.")
    }

    err := e.New("an error happened")
    l.Errorf(nil, err, "Error message")
    if tla.errorTriggered == false {
        t.Error("Error message not getting through.")
    }
}

func TestStaticConfiguration(t *testing.T) {
    cp := NewStaticConfigurationProvider()
    scp := cp.(*StaticConfigurationProvider)

    scp.SetFormat("aa")
    scp.SetAdapterName("bb")
    scp.SetLevelName("cc")
    scp.SetIncludeNouns("dd")
    scp.SetExcludeNouns("ee")
    scp.SetExcludeBypassLevelName("ff")

    LoadConfiguration(scp)

    if format != "aa" {
        t.Error("Static configuration provider was not set correctly: format")
    }

    if GetDefaultAdapterName() != "bb" {
        t.Error("Static configuration provider was not set correctly: adapterName")
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
        t.Error("Adapter on logger when no adapters was not nil.")
    }

    // Should be allowed, but nothing will happen.
    err := e.New("an error happened")
    l.Errorf(nil, err, "Error message")
}
