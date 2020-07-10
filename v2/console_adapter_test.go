package log

import (
    "testing"
)

func TestConsole(t *testing.T) {
    ecp := NewEnvironmentConfigurationProvider()
    LoadConfiguration(ecp)

    ClearAdapters()

    cla := NewConsoleLogAdapter()
    AddAdapter("console", cla)

    an := GetDefaultAdapterName()
    if an != "console" {
        t.Error("Console adapter was not properly registered.")
    }

    l := NewLoggerWithAdapterName("consoleTest", "console")
    l.Debugf(nil, "")

    if l.Adapter() == nil {
        t.Error("Adapter wasn't initialized correctly.")
    }
}
