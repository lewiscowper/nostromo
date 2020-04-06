package task

import (
	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/pokanop/nostromo/prompt"
	"github.com/pokanop/nostromo/shell"
	"github.com/pokanop/nostromo/stringutil"
	"github.com/pokanop/nostromo/version"
	"github.com/shivamMg/ppds/tree"
	"strings"
)

var ver *version.Info

// SetVersion should be called before any task to ensure manifest is updated
func SetVersion(v *version.Info) {
	ver = v
}

// InitConfig of nostromo config file if not already initialized
func InitConfig() {
	cfg := checkConfigQuiet()

	if cfg == nil {
		cfg = config.NewConfig(config.Path, model.NewManifest())
		err := pathutil.EnsurePath("~/.nostromo")
		if err != nil {
			log.Error(err)
			return
		}

		log.Highlight("nostromo config created")
	} else {
		log.Highlight("nostromo config exists, updating")
	}

	err := saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// DestroyConfig deletes nostromo config file
func DestroyConfig() {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Delete()
	if err != nil {
		log.Error(err)
		return
	}

	log.Highlight("nostromo config deleted")

	err = shell.Commit(model.NewManifest())
	if err != nil {
		log.Error(err)
		return
	}
}

// ShowConfig for nostromo config file
func ShowConfig(asJSON bool, asYAML bool, asTree bool) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	if asJSON || asYAML {
		log.Bold("[manifest]")
		if asJSON {
			log.Regular(m.AsJSON())
			log.Regular()
		} else if asYAML {
			log.Regular(m.AsYAML())
		}
	} else if asTree {
		tree.PrintHr(m)
	} else {
		log.Bold("[manifest]")
		logFields(m, m.Config.Verbose)

		log.Bold("\n[config]")
		logFields(m.Config, m.Config.Verbose)

		if len(m.Commands) > 0 {
			log.Bold("\n[commands]")
			for _, cmd := range m.Commands {
				cmd.Walk(func(c *model.Command, s *bool) {
					logFields(c, m.Config.Verbose)
					if m.Config.Verbose {
						log.Regular()
					}
				})
			}
		} else if m.Config.Verbose {
			log.Regular()
		}

		if !m.Config.Verbose {
			log.Regular()
		}
	}

	lines, err := shell.InitFileLines()
	if err != nil {
		return
	}

	log.Bold("[profile]")
	if len(lines) > 0 {
		log.Regular(strings.TrimSpace(lines))
	} else {
		log.Regular("empty")
	}
}

// SetConfig updates properties for nostromo settings
func SetConfig(key, value string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Set(key, value)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// GetConfig reads properties from nostromo settings
func GetConfig(key string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	log.Highlight(cfg.Get(key))
}

// AddInteractive adds a command or substitution through user prompts
func AddInteractive() {
	isCmd := prompt.Choose("Choose what you would like to add", []string{"command", "substitution"}) == 0
	keypath := prompt.String("Enter a key path to attach your command (root)")
	if isCmd {
		cmd := prompt.StringRequired("The actual command to run")
		alias := prompt.StringRequired("The alias or shortcut to use")
		description := prompt.String("Provide a description for your command")
		languages := shell.SupportedLanguages()
		language := languages[prompt.Choose("Choose a language to use (sh)", languages)]
		var snippet string
		if language != "sh" {
			snippet = prompt.StringRequired("Provide the code snippet to run")
		}
		aliasOnly := prompt.Confirm("Is this command a standard alias")
		var mode string
		if !aliasOnly {
			modes := model.SupportedModes()
			mode = modes[prompt.Choose("Choose a command mode to use (concatenate)", modes)]
		}
		if len(keypath) == 0 {
			keypath = alias
		} else {
			keypath = strings.Join([]string{keypath, alias}, ".")
		}
		AddCommand(keypath, cmd, description, snippet, language, aliasOnly, mode)
	} else {
		sub := prompt.StringRequired("Original value to replace")
		alias := prompt.StringRequired("Substitution to use")
		AddSubstitution(keypath, sub, alias)
	}
}

// AddCommand to the manifest
func AddCommand(keyPath, command, description, code, language string, aliasOnly bool, mode string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	snippet := &model.Code{
		Language: language,
		Snippet:  code,
	}

	err := m.AddCommand(keyPath, command, description, snippet, aliasOnly, mode)
	if err != nil {
		log.Error(err)
		return
	}

	cmd := m.Find(keyPath)
	if cmd == nil {
		log.Error("unable to find newly created command")
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}

	logFields(cmd, m.Config.Verbose)
}

// RemoveCommand from the manifest
func RemoveCommand(keyPath string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest().RemoveCommand(keyPath)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// AddSubstitution to the manifest
func AddSubstitution(keyPath, name, alias string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	err := m.AddSubstitution(keyPath, name, alias)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
		return
	}

	logFields(m.Find(keyPath), m.Config.Verbose)
}

// RemoveSubstitution from the manifest
func RemoveSubstitution(keyPath, alias string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest().RemoveSubstitution(keyPath, alias)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// EvalString returns a command that can be used with `eval`
func EvalString(args []string) {
	log.SetEcho(true)

	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	language, cmd, err := m.ExecutionString(stringutil.SanitizeArgs(args))
	if err != nil {
		log.Error(err)
		return
	}

	cmdStr, err := shell.EvalString(cmd, language, m.Config.Verbose)
	if err != nil {
		log.Error(err)
	}

	log.Print(cmdStr)
}

// Find matching commands and substitutions
func Find(name string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	matchingCmds := []*model.Command{}
	matchingSubs := []*model.Command{}

	for _, cmd := range m.Commands {
		cmd.Walk(func(c *model.Command, s *bool) {
			if stringutil.ContainsCaseInsensitive(c.Name, name) || stringutil.ContainsCaseInsensitive(c.Alias, name) {
				matchingCmds = append(matchingCmds, c)
			}
			for _, sub := range c.Subs {
				if stringutil.ContainsCaseInsensitive(sub.Name, name) || stringutil.ContainsCaseInsensitive(sub.Alias, name) {
					matchingSubs = append(matchingSubs, c)
				}
			}
		})
	}

	if len(matchingCmds) == 0 && len(matchingSubs) == 0 {
		log.Highlight("no matching commands or substitutions found")
		return
	}

	log.Regular("[commands]")
	for _, cmd := range matchingCmds {
		logFields(cmd, m.Config.Verbose)
		if m.Config.Verbose {
			log.Regular()
		}
	}

	if !m.Config.Verbose {
		log.Regular()
	}
	log.Regular("[substitutions]")
	for _, cmd := range matchingSubs {
		logFields(cmd, m.Config.Verbose)
		if m.Config.Verbose {
			log.Regular()
		}
	}
}

func checkConfigQuiet() *config.Config {
	return checkConfigCommon(true)
}

func checkConfig() *config.Config {
	return checkConfigCommon(false)
}

func checkConfigCommon(quiet bool) *config.Config {
	cfg, err := config.Parse(config.Path)
	if err != nil {
		if !quiet {
			log.Error(err)
			log.Info("unable to open config file, be sure to run `nostromo init` if you haven't already")
		}
		return nil
	}

	log.SetVerbose(cfg.Manifest().Config.Verbose)

	return cfg
}

func saveConfig(cfg *config.Config) error {
	m := cfg.Manifest()
	m.Version = ver.SemVer

	err := cfg.Save()
	if err != nil {
		return err
	}

	err = shell.Commit(m)
	if err != nil {
		return err
	}

	return nil
}

func logFields(mapper log.FieldMapper, verbose bool) {
	if verbose {
		log.Table(mapper)
		return
	}
	log.Fields(mapper)
}
