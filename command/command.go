package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bowery/prompt"
	"github.com/labstack/gommon/color"
	"github.com/mkideal/cli"
	"github.com/mkideal/onepw/core"
	"github.com/mkideal/pkg/debug"
	"github.com/mkideal/pkg/textutil"
)

func Exec(args []string) error {
	return root.Run(args)
}

func init() {
	root = cli.Root(root,
		cli.Tree(help),
		cli.Tree(version),
		cli.Tree(initCmd),
		cli.Tree(add),
		cli.Tree(remove),
		cli.Tree(list),
		cli.Tree(find),
		cli.Tree(upgrade),
		cli.Tree(info),
	)
}

//--------
// Config
//--------

// Configure ...
type Configure interface {
	Filename() string
	MasterPassword() string
	Debug() bool
}

// Config implementes Configure interface, represents onepw config
type Config struct {
	Master      string `pw:"master" usage:"Your master password" dft:"$ONEPW_MASTER" prompt:"Type the master password"`
	EnableDebug bool   `cli:"debug" usage:"Enable debug mode" dft:"false"`
}

// Filename returns password data filename
func (cfg Config) Filename() string {
	filename := os.Getenv("ONEPW_FILE")
	if filename == "" {
		filename = "password.data"
	}
	return filename
}

// MasterPassword returns master password
func (cfg Config) MasterPassword() string {
	return cfg.Master
}

// Debug returns debug mode
func (cfg Config) Debug() bool {
	return cfg.EnableDebug
}

var box *core.Box

//--------------
// root command
//--------------

type rootT struct {
	cli.Helper2
	Version bool `cli:"!v,version" usage:"Display version information"`
}

var root = &cli.Command{
	Name: os.Args[0],
	Desc: textutil.Tpl("{{.onepw}} is a command line tool for managing passwords, open-source on {{.repo}}", map[string]string{
		"onepw": color.Bold("onepw"),
		"repo":  color.Blue("https://github.com/mkideal/onepw"),
	}),
	Text: textutil.Tpl(`{{.usage}}: {{.onepw}} <COMMAND> [OPTIONS]`, map[string]string{
		"onepw": color.Bold("onepw"),
		"usage": color.Bold("Usage"),
	}),
	Argv:   func() interface{} { return new(rootT) },
	NumArg: cli.AtLeast(1),

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*rootT)
		if argv.Version {
			ctx.String("%s\n", appVersion)
			return cli.ExitError
		}
		return nil
	},

	OnRootBefore: func(ctx *cli.Context) error {
		if argv := ctx.Argv(); argv != nil {
			if t, ok := argv.(Configure); ok {
				debug.Switch(t.Debug())
				repo := core.NewFileRepository(t.Filename())
				box = core.NewBox(repo)
				if t.MasterPassword() != "" {
					return box.Init(t.MasterPassword())
				}
				return nil
			}
		}
		return fmt.Errorf("box is nil")
	},

	Fn: func(ctx *cli.Context) error {
		return nil
	},
}

//--------------
// help command
//--------------

var help = cli.HelpCommand("Display help information")

//-----------------
// version command
//-----------------

var version = &cli.Command{
	Name:   "version",
	Desc:   "Display version information",
	NoHook: true,

	Fn: func(ctx *cli.Context) error {
		ctx.String(appVersion + "\n")
		return nil
	},
}

//--------------
// init command
//--------------
type initT struct {
	cli.Helper2
	Config
	Update bool `cli:"u,update" usage:"Whether to update the master password" dft:"false"`
}

func (argv *initT) Validate(ctx *cli.Context) error {
	if argv.Filename() == "" {
		return fmt.Errorf("FILE is empty")
	}
	return nil
}

var initCmd = &cli.Command{
	Name: "init",
	Desc: "Init password box or change the master password",
	Argv: func() interface{} { return new(initT) },

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*initT)
		if argv.Update {
			return nil
		}
		cpw, err := prompt.Password("Repeat the master password: ")
		if err != nil {
			return err
		}
		if argv.Master != string(cpw) {
			return fmt.Errorf(ctx.Color().Red("master password mismatch"))
		}

		if _, err := os.Lstat(argv.Filename()); err != nil {
			if os.IsNotExist(err) {
				dir, _ := filepath.Split(argv.Filename())
				if dir != "" && dir != "." {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return err
					}
				}
				file, err := os.Create(argv.Filename())
				if err != nil {
					return err
				}
				file.Close()
			}
		}
		return nil
	},

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*initT)
		if argv.Update {
			pw, err := prompt.Password("Type a new master password:")
			if err != nil {
				return err
			}
			cpw, err := prompt.Password("Repeat the new master password:")
			if err != nil {
				return err
			}
			if string(pw) != string(cpw) {
				return fmt.Errorf(ctx.Color().Red("new master password mismatch"))
			}
			return box.Update(string(pw))
		}
		return nil
	},
}

//-------------
// add command
//-------------
type addT struct {
	cli.Helper2
	Config
	core.Password
	Pw  string `pw:"p,password" usage:"The password you decided to use" name:"PASSWORD" prompt:"Type the password"`
	Cpw string `pw:"C,confirm" usage:"Confirm password which must be same as PASSWORD" prompt:"Repeat the password"`
}

func (argv *addT) Validate(ctx *cli.Context) error {
	if argv.Pw != argv.Cpw {
		return fmt.Errorf("2 passwords mismatch")
	}
	return core.CheckPassword(argv.Pw)
}

var add = &cli.Command{
	Name: "add",
	Desc: "Add a new password or update(while with --id parameter) the old password",
	Argv: func() interface{} {
		argv := new(addT)
		argv.Password = *core.NewEmptyPassword()
		return argv
	},

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*addT)
		argv.Password.PlainPassword = argv.Pw
		id, new, err := box.Add(&argv.Password)
		if err != nil {
			return err
		}
		if new {
			ctx.String("password %s added\n", ctx.Color().Cyan(id))
		} else {
			ctx.String("password %s updated\n", ctx.Color().Cyan(id))
		}
		return nil
	},
}

//--------
// remove
//--------

type removeT struct {
	cli.Helper2
	Config
	All bool `cli:"a,all" usage:"Remove all found passwords" dft:"false"`
}

var remove = &cli.Command{
	Name:        "remove",
	Aliases:     []string{"rm", "del", "delete"},
	Desc:        "Remove passwords by IDs or (category,account)",
	Text:        "Usage: onepw rm [IDs...] [OPTIONS]",
	Argv:        func() interface{} { return new(removeT) },
	CanSubRoute: true,

	Fn: func(ctx *cli.Context) error {
		var (
			argv       = ctx.Argv().(*removeT)
			deletedIds []string
			err        error
			ids        = ctx.Args()
		)
		if len(ids) > 0 {
			deletedIds, err = box.Remove(ids, argv.All)
		} else if argv.All {
			deletedIds, err = box.Clear()
		}

		if err != nil {
			return err
		}
		ctx.String("deleted passwords:\n")
		ctx.String(ctx.Color().Cyan(strings.Join(deletedIds, "\n")))
		ctx.String("\n")
		return nil
	},
}

//------
// list
//------

type listT struct {
	cli.Helper2
	Config
	NoHeader   bool `cli:"no-header" usage:"Don't print header line" dft:"false"`
	ShowHidden bool `cli:"H,hidden" usage:"Whether to list hidden passwords"`
}

var list = &cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Desc:    "List all passwords",
	Argv:    func() interface{} { return new(listT) },

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*listT)
		return box.List(ctx, argv.NoHeader, argv.ShowHidden)
	},
}

//--------------
// find command
//--------------

type findT struct {
	cli.Helper2
	Config
	JustPassword bool `cli:"p,just-password" usage:"Just show password" dft:"false"`
	JustFirst    bool `cli:"f,just-first" usage:"Just show first result" dft:"false"`
}

var find = &cli.Command{
	Name:        "find",
	Desc:        "Find password by ID,category,account,tag or site and so on",
	Text:        "Usage: onepw find <WORD>",
	Argv:        func() interface{} { return new(findT) },
	CanSubRoute: true,

	OnBefore: func(ctx *cli.Context) error {
		if len(ctx.Args()) != 1 {
			ctx.WriteUsage()
			return cli.ExitError
		}
		return nil
	},

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*findT)
		box.Find(ctx, ctx.Args()[0], argv.JustPassword, argv.JustFirst)
		return nil
	},
}

//-----------------
// upgrade command
//-----------------

type upgradeT struct {
	cli.Helper2
	Config
}

var upgrade = &cli.Command{
	Name:    "upgrade",
	Aliases: []string{"up"},
	Desc:    "Upgrade to newest version",
	Argv:    func() interface{} { return new(upgradeT) },

	Fn: func(ctx *cli.Context) error {
		from, to, err := box.Upgrade()
		if err != nil {
			return err
		}
		ctx.String("upgrade from %d to %d!\n", from, to)
		return nil
	},
}

//--------------
// info command
//--------------
type infoT struct {
	cli.Helper2
	Config
	All bool `cli:"a,all" usage:"show all found passwords"`
}

var info = &cli.Command{
	Name:        "show",
	Aliases:     []string{"info"},
	Desc:        "Show low-level information of password",
	Text:        "Usage: onepw show <IDs...>",
	Argv:        func() interface{} { return new(infoT) },
	CanSubRoute: true,
	NumArg:      cli.AtLeast(1),

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*infoT)
		return box.Inspect(ctx, ctx.Args(), argv.All)
	},
}
