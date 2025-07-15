package plugin

type Plugin interface {
	Name() string
	Install(version string) error
	Uninstall(version string) error
	List() ([]string, error)
	ListRemote() ([]string, error)
	Use(version string) error
	Global(version string) error
	Local(version string, projectDir string) error
}

var registry = map[string]Plugin{}

func Register(lang string, p Plugin) {
	registry[lang] = p
}

func Get(lang string) (Plugin, bool) {
	p, ok := registry[lang]
	return p, ok
}

// All returns all registered plugins as a map[lang]Plugin
func All() map[string]Plugin {
	return registry
}
