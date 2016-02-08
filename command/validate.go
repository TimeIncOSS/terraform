package command

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform/config"
)

// ValidateCommand is a Command implementation that validates the terraform files
type ValidateCommand struct {
	Meta
}

const defaultPath = "."

func (c *ValidateCommand) Help() string {
	helpText := `
Usage: terraform validate [dir]

  Syntactically validates the Terraform files

Options:

  -destroy      If set, terraform will validate configuration
                in the context of destroying resources.
`
	return strings.TrimSpace(helpText)
}

func (c *ValidateCommand) Run(args []string) int {
	args = c.Meta.process(args, false)

	var dirPath string
	var destroy bool

	cmdFlags := c.Meta.flagSet("validate")
	cmdFlags.BoolVar(&destroy, "destroy", false, "destroy")

	if len(args) == 1 {
		dirPath = args[0]
	} else {
		dirPath = "."
	}
	dir, err := filepath.Abs(dirPath)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Unable to locate directory %v\n", err.Error()))
	}

	rtnCode := c.validate(dir, destroy)

	return rtnCode
}

func (c *ValidateCommand) Synopsis() string {
	return "Validates the Terraform files"
}

func (c *ValidateCommand) validate(dir string, destroy bool) int {
	cfg, err := config.LoadDir(dir)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Loading files failed: %v\n", err.Error()))
		return 1
	}

	err = cfg.Validate()
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Validation failed: %v\n", err.Error()))
		return 1
	}

	ctx, _, err := c.Context(contextOpts{
		Destroy:     destroy,
		Path:        dir,
		StatePath:   c.Meta.statePath,
		Parallelism: c.Meta.parallelism,
	})

	ws, es := ctx.Validate()

	if len(ws) > 0 {
		c.Ui.Warn("Warnings:\n")
		for _, w := range ws {
			c.Ui.Warn(fmt.Sprintf("  * %s", w))
		}

		if len(es) > 0 {
			c.Ui.Output("")
		}
	}

	if len(es) > 0 {
		c.Ui.Error("Errors:\n")
		for _, e := range es {
			c.Ui.Error(fmt.Sprintf("  * %s", e))
		}
	}

	if len(ws) > 0 || len(es) > 0 {
		return 1
	}

	return 0
}
