package log

import (
    "os"
    "fmt"
)

// Config keys.
const (
    ckFormat = "LogFormat"
    ckAdapterName = "LogAdapterName"
    ckLevelName = "LogLevelName"
    ckIncludeNouns = "LogIncludeNouns"
    ckExcludeNouns = "LogExcludeNouns"
    ckExcludeBypassLevelName = "LogExcludeBypassLevelName"
)

// Other constants
const (
    defaultFormat = "{{.Noun}}:{{if eq .ExcludeBypass true}} [BYPASS]{{end}} {{.Message}}"
    defaultLevelName = LevelNameInfo
)

// Config
var (
    // Alternative format (defaults to .
    format = defaultFormat

    // Alternative adapter (defaults to "appengine").
    adapterName = ""

    // Alternative level at which to display log-items (defaults to 
    // "info").
    levelName = defaultLevelName

    // Configuration-driven comma-separated list of nouns to include. Defaults
    // to empty.
    includeNouns = ""

    // Configuration-driven comma-separated list of nouns to exclude. Defaults 
    // to empty.
    excludeNouns = ""

    // Level at which to disregard exclusion (if the severity of a message 
    // meets or exceed this, always display). Defaults to empty.
    excludeBypassLevelName = ""
)

// Other
var (
    configurationLoaded = false
)

// Return the current default adapter name.
func GetDefaultAdapterName() string {
    return adapterName
}

// The adapter will automatically be the first one registered. This overrides 
// that.
func SetDefaultAdapterName(name string) {
    adapterName = name
}

func LoadConfiguration(cp ConfigurationProvider) {
    adapterName = cp.AdapterName()
    includeNouns = cp.IncludeNouns()
    excludeNouns = cp.ExcludeNouns()
    excludeBypassLevelName = cp.ExcludeBypassLevelName()

    f := cp.Format()
    if f != "" {
        format = f
    }

    ln := cp.LevelName()
    if ln != "" {
        levelName = ln
    }

    configurationLoaded = true
}

func GetConfigDump() string {
    return fmt.Sprintf(
        "Current configuration:\n" +
        "  FORMAT=[%s]\n" +
        "  ADAPTER-NAME=[%s]\n" +
        "  LEVEL-NAME=[%s]\n" +
        "  INCLUDE-NOUNS=[%s]\n" +
        "  EXCLUDE-NOUNS=[%s]\n" +
        "  EXCLUDE-BYPASS-LEVEL-NAME=[%s]", 
        format, adapterName, levelName, includeNouns, excludeNouns, excludeBypassLevelName)
}

func IsConfigurationLoaded() bool {
    return configurationLoaded
}

type ConfigurationProvider interface {
    // Alternative format (defaults to .
    Format() string

    // Alternative adapter (defaults to "appengine").
    AdapterName() string

    // Alternative level at which to display log-items (defaults to 
    // "info").
    LevelName() string

    // Configuration-driven comma-separated list of nouns to include. Defaults
    // to empty.
    IncludeNouns() string

    // Configuration-driven comma-separated list of nouns to exclude. Defaults 
    // to empty.
    ExcludeNouns() string

    // Level at which to disregard exclusion (if the severity of a message 
    // meets or exceed this, always display). Defaults to empty.
    ExcludeBypassLevelName() string
}

// Environment configuration-provider.
type EnvironmentConfigurationProvider struct {

}

func NewEnvironmentConfigurationProvider() ConfigurationProvider {
    return new(EnvironmentConfigurationProvider)
}

func (ecp *EnvironmentConfigurationProvider) Format() string {
    return os.Getenv(ckFormat)
}

func (ecp *EnvironmentConfigurationProvider) AdapterName() string {
    return os.Getenv(ckAdapterName)
}

func (ecp *EnvironmentConfigurationProvider) LevelName() string {
    return os.Getenv(ckLevelName)
}

func (ecp *EnvironmentConfigurationProvider) IncludeNouns() string {
    return os.Getenv(ckIncludeNouns)
}

func (ecp *EnvironmentConfigurationProvider) ExcludeNouns() string {
    return os.Getenv(ckExcludeNouns)
}

func (ecp *EnvironmentConfigurationProvider) ExcludeBypassLevelName() string {
    return os.Getenv(ckExcludeBypassLevelName)
}

// Static configuration-provider.
type StaticConfigurationProvider struct {
    format string
    adapterName string
    levelName string
    includeNouns string
    excludeNouns string
    excludeBypassLevelName string
}

func NewStaticConfigurationProvider() ConfigurationProvider {
    return new(StaticConfigurationProvider)
}

func (scp *StaticConfigurationProvider) SetFormat(format string) {
    scp.format = format
}

func (scp *StaticConfigurationProvider) SetAdapterName(adapterName string) {
    scp.adapterName = adapterName
}

func (scp *StaticConfigurationProvider) SetLevelName(levelName string) {
    scp.levelName = levelName
}

func (scp *StaticConfigurationProvider) SetIncludeNouns(includeNouns string) {
    scp.includeNouns = includeNouns
}

func (scp *StaticConfigurationProvider) SetExcludeNouns(excludeNouns string) {
    scp.excludeNouns = excludeNouns
}

func (scp *StaticConfigurationProvider) SetExcludeBypassLevelName(excludeBypassLevelName string) {
    scp.excludeBypassLevelName = excludeBypassLevelName
}


func (scp *StaticConfigurationProvider) Format() string {
    return scp.format
}

func (scp *StaticConfigurationProvider) AdapterName() string {
    return scp.adapterName
}

func (scp *StaticConfigurationProvider) LevelName() string {
    return scp.levelName
}

func (scp *StaticConfigurationProvider) IncludeNouns() string {
    return scp.includeNouns
}

func (scp *StaticConfigurationProvider) ExcludeNouns() string {
    return scp.excludeNouns
}

func (scp *StaticConfigurationProvider) ExcludeBypassLevelName() string {
    return scp.excludeBypassLevelName
}

func init() {
    // Do the initial configuration-load from the environment. We gotta seed it
    // with something for simplicity's sake.
    ecp := NewEnvironmentConfigurationProvider()
    LoadConfiguration(ecp)
}
