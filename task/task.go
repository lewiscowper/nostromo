package task

import (
	"strings"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/pokanop/nostromo/shell"
)

// InitConfig of nostromo config file if not already initialized
func InitConfig() {
	cfg := checkConfigQuiet()
	if cfg == nil {
		cfg = config.NewConfig(config.ConfigPath, model.NewManifest())
		err := pathutil.EnsurePath("~/.nostromo")
		if err != nil {
			log.Error(err)
			return
		}
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

	err = shell.Commit(model.NewManifest())
	if err != nil {
		log.Error(err)
		return
	}
}

// ShowConfig for nostromo config file
func ShowConfig(raw bool) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	if raw {
		log.Highlight("[raw json]")
		log.Regular(cfg.Manifest.AsJSON())
		lines, err := shell.InitFileLines()
		if err != nil {
			return
		}

		log.Highlight("\n[profile]")
		log.Regular(strings.TrimSpace(lines))
	} else {
		log.Regular("[manifest]")
		log.Fields(cfg.Manifest)

		log.Regular("\n[config]")
		log.Fields(cfg.Manifest.Config)

		if len(cfg.Manifest.Commands) > 0 {
			log.Regular("\n[commands]")
			for _, cmd := range cfg.Manifest.Commands {
				cmd.Walk(func(c *model.Command, s *bool) {
					log.Fields(c)
					if cfg.Manifest.Config.Verbose {
						log.Regular()
					}
				})
			}
		}
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

// AddCommand to the manifest
func AddCommand(keyPath, command, description string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.AddCommand(keyPath, command, description)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}

	log.Fields(cfg.Manifest.Find(keyPath))
}

// RemoveCommand from the manifest
func RemoveCommand(keyPath string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.RemoveCommand(keyPath)
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

	err := cfg.Manifest.AddSubstitution(keyPath, name, alias)
	if err != nil {
		log.Error(err)
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}

	log.Fields(cfg.Manifest.Find(keyPath))
}

// RemoveSubstitution from the manifest
func RemoveSubstitution(keyPath, alias string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest.RemoveSubstitution(keyPath, alias)
	if err != nil {
		log.Error(err)
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// Run a command from the manifest
func Run(args []string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	cmd, err := cfg.Manifest.ExecutionString(sanitizeArgs(args))
	if err != nil {
		log.Error(err)
		return
	}

	err = shell.Run(cmd, cfg.Manifest.Config.Verbose)
	if err != nil {
		log.Error(err)
	}
}

// Find matching commands and substitutions
func Find(name string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	matchingCmds := []*model.Command{}
	matchingSubs := []*model.Command{}

	for _, cmd := range cfg.Manifest.Commands {
		cmd.Walk(func(c *model.Command, s *bool) {
			if containsCaseInsensitive(c.Name, name) || containsCaseInsensitive(c.Alias, name) {
				matchingCmds = append(matchingCmds, c)
			}
			for _, sub := range c.Subs {
				if containsCaseInsensitive(sub.Name, name) || containsCaseInsensitive(sub.Alias, name) {
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
		log.Fields(cmd)
		if cfg.Manifest.Config.Verbose {
			log.Regular()
		}
	}

	if !cfg.Manifest.Config.Verbose {
		log.Regular()
	}
	log.Regular("[substitutions]")
	for _, cmd := range matchingSubs {
		log.Fields(cmd)
		if cfg.Manifest.Config.Verbose {
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
	cfg, err := config.Parse(config.ConfigPath)
	if err != nil {
		if !quiet {
			log.Error(err)
		}
		return nil
	}

	log.SetOptions(cfg.Manifest.Config.Verbose)

	return cfg
}

func saveConfig(cfg *config.Config) error {
	err := cfg.Save()
	if err != nil {
		return err
	}

	err = shell.Commit(cfg.Manifest)
	if err != nil {
		return err
	}

	return nil
}

func sanitizeArgs(args []string) []string {
	sanitizedArgs := []string{}
	for _, arg := range args {
		if len(arg) > 0 {
			sanitizedArgs = append(sanitizedArgs, strings.TrimSpace(arg))
		}
	}
	return sanitizedArgs
}

func containsCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
