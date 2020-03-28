package spec

// LanguageTemplate read from template.yml within root of a language template folder
type LanguageTemplate struct {
	Language     string        `yaml:"language,omitempty"`
	BuildOptions []BuildOption `yaml:"build_options,omitempty"`
	// WelcomeMessage is printed to the user after generating a function
	WelcomeMessage string `yaml:"welcome_message,omitempty"`
	// HandlerFolder to copy the function code into
	HandlerFolder string `yaml:"handler_folder,omitempty"`
}

// BuildOption a named build option for one or more packages
type BuildOption struct {
	Name     string   `yaml:"name"`
	Packages []string `yaml:"packages"`
}
