package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Bowery/prompt"
	"github.com/labstack/gommon/color"
	"github.com/mattn/go-colorable"
	"github.com/mkideal/cli"
	"github.com/mkideal/onepw/core"
	"github.com/mkideal/pkg/debug"
	"github.com/mkideal/pkg/textutil"
)

func main() {
	cli.SetUsageStyle(cli.ManualStyle)
	if err := cli.Root(root,
		cli.Tree(help),
		cli.Tree(version),
		cli.Tree(initCmd),
		cli.Tree(add),
		cli.Tree(remove),
		cli.Tree(list),
		cli.Tree(find),
		cli.Tree(upgrade),
		cli.Tree(generate),
		cli.Tree(info),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(colorable.NewColorableStderr(), err)
		os.Exit(1)
	}
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
	Master      string `pw:"master" usage:"master password" dft:"$ONEPW_MASTER" prompt:"type the master password"`
	EnableDebug bool   `cli:"debug" usage:"usage debug mode" dft:"false"`
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
	cli.Helper
	Version bool `cli:"!v,version" usage:"display version information"`
}

var root = &cli.Command{
	Name: os.Args[0],
	Desc: textutil.Tpl("{{.onepw}} is a command line tool for managing passwords, open-source on {{.repo}}", map[string]string{
		"onepw": color.Bold("onepw"),
		"repo":  color.Blue("https://github.com/mkideal/onepw"),
	}),
	Text: textutil.Tpl(`{{.usage}}: {{.onepw}} <COMMAND> [OPTIONS]

{{.basicworkflow}}:

	#1. init, create file password.data
	$> {{.onepw}} init

	#2. add a new password
	$> {{.onepw}} add -c=email -u user@example.com
	type the password:
	repeat the password:

	#3. list all passwords
	$> {{.onepw}} ls

	#optional
	# upload cloud(e.g. dropbox or github or bitbucket ...)`, map[string]string{
		"onepw":         color.Bold("onepw"),
		"usage":         color.Bold("Usage"),
		"basicworkflow": color.Bold("Basic workflow"),
	}),
	Argv: func() interface{} { return new(rootT) },

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*rootT)
		if argv.Help || len(ctx.NativeArgs()) == 0 {
			ctx.WriteUsage()
			return cli.ExitError
		}
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

var help = cli.HelpCommand("display help")

//-----------------
// version command
//-----------------

var version = &cli.Command{
	Name:   "version",
	Desc:   "display version",
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
	cli.Helper
	Config
	Update bool `cli:"u,update" usage:"whether update master password" dft:"false"`
}

func (argv *initT) Validate(ctx *cli.Context) error {
	if argv.Filename() == "" {
		return fmt.Errorf("FILE is empty")
	}
	return nil
}

var initCmd = &cli.Command{
	Name: "init",
	Desc: "init password box or modify master password",
	Argv: func() interface{} { return new(initT) },

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*initT)
		if argv.Help {
			ctx.WriteUsage()
			return cli.ExitError
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
			data, err := prompt.Password("type the new master password: ")
			if err != nil {
				return err
			}
			return box.Update(string(data))
		}
		return nil
	},
}

//-------------
// add command
//-------------
type addT struct {
	cli.Helper
	Config
	core.Password
	Pw  string `pw:"pw,password" usage:"the password" prompt:"type the password"`
	Cpw string `pw:"cpw,confirm-password" usage:"confirm password" prompt:"repeat the password"`
}

func (argv *addT) Validate(ctx *cli.Context) error {
	if argv.Pw != argv.Cpw {
		return fmt.Errorf("2 passwords mismatch")
	}
	return core.CheckPassword(argv.Pw)
}

var add = &cli.Command{
	Name: "add",
	Desc: "add a new password or update old password",
	Argv: func() interface{} {
		argv := new(addT)
		argv.Password = *core.NewEmptyPassword()
		return argv
	},

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*addT)
		if argv.Help {
			ctx.WriteUsage()
			return cli.ExitError
		}
		return nil
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
	cli.Helper
	Config
	All bool `cli:"a,all" usage:"remove all found passwords" dft:"false"`
}

var remove = &cli.Command{
	Name:        "remove",
	Aliases:     []string{"rm", "del", "delete"},
	Desc:        "remove passwords by ids or (category,account)",
	Text:        "Usage: onepw rm [ids...] [OPTIONS]",
	Argv:        func() interface{} { return new(removeT) },
	CanSubRoute: true,

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*removeT)
		if argv.Help {
			ctx.WriteUsage()
			return cli.ExitError
		}
		return nil
	},

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
	cli.Helper
	Config
	NoHeader bool `cli:"no-header" usage:"don't print header line" dft:"false"`
}

var list = &cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Desc:    "list all passwords",
	Argv:    func() interface{} { return new(listT) },

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*listT)
		if argv.Help {
			ctx.WriteUsage()
			return cli.ExitError
		}
		return nil
	},

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*listT)
		return box.List(ctx, argv.NoHeader)
	},
}

//--------------
// find command
//--------------

type findT struct {
	cli.Helper
	Config
	JustPassword bool `cli:"p,just-password" usage:"only show password" dft:"false"`
	JustFirst    bool `cli:"f,just-first" usage:"only show first result" dft:"false"`
}

var find = &cli.Command{
	Name:        "find",
	Desc:        "find password by id,category,account,tag,site and so on",
	Text:        "Usage: onepw find <WORD>",
	Argv:        func() interface{} { return new(findT) },
	CanSubRoute: true,

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*findT)
		if argv.Help || len(ctx.Args()) != 1 {
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
	cli.Helper
	Config
}

var upgrade = &cli.Command{
	Name:    "upgrade",
	Aliases: []string{"up"},
	Desc:    "upgrade to newest version",
	Argv:    func() interface{} { return new(upgradeT) },

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*upgradeT)
		if argv.Help || len(ctx.NativeArgs()) != 0 {
			ctx.WriteUsage()
			return cli.ExitError
		}
		return nil
	},

	Fn: func(ctx *cli.Context) error {
		from, to, err := box.Upgrade()
		if err != nil {
			return err
		}
		ctx.String("upgrade from %d to %d!\n", from, to)
		return nil
	},
}

//------------------
// generate command
//------------------
type generateT struct {
	cli.Helper
	Num                uint16 `cli:"n,number" usage:"number of generated passwords" dft:"1" name:"N"`
	ContainDigit       bool   `cli:"d,digit" usage:"whether the password contains digit" dft:"false"`
	ContainLowerChar   bool   `cli:"c,lower-char" usage:"whether the password contains lowercase character" dft:"false"`
	ContainUpperChar   bool   `cli:"C,upper-char" usage:"whether the password contains uppercase character" dft:"false"`
	ContainSpecialChar bool   `cli:"s,special-char" usage:"whether the password contains the special character" dft:"false"`
	SpecialCharSet     string `cli:"sset,special-set" usage:"custom special character set"`

	length int `cli:"-"`
}

var (
	digits         = []byte("0123456789")
	lowercaseChars = []byte("abcdefghijklmnopqrstuvwxyz")
	uppercaseChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	specialChars   = []byte("~!@#$%^&*")
)

func (argv *generateT) Validate(ctx *cli.Context) error {
	if argv.Num == 0 {
		return fmt.Errorf("N must > 0")
	}
	args := ctx.Args()
	if len(args) == 1 {
		length, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("LEN must be a number")
		}
		if length <= 0 {
			return fmt.Errorf("LEN must > 0")
		}
		argv.length = length
	}

	return nil
}

var generate = &cli.Command{
	Name:        "generate",
	Aliases:     []string{"gen"},
	Desc:        "a utility command for generating password",
	Text:        "Usage: onepw generate [OPTIONS] LEN",
	Argv:        func() interface{} { return new(generateT) },
	NoHook:      true,
	CanSubRoute: true,

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*generateT)
		if argv.Help || len(ctx.Args()) != 1 {
			ctx.WriteUsage()
			return nil
		}
		if !argv.ContainDigit && !argv.ContainLowerChar && !argv.ContainUpperChar && !argv.ContainSpecialChar {
			argv.ContainDigit = true
			argv.ContainLowerChar = true
			argv.ContainUpperChar = true
		}
		rand.Seed(time.Now().UnixNano())

		charSetBuff := bytes.NewBufferString("")
		if argv.ContainDigit {
			charSetBuff.Write(digits)
		}
		if argv.ContainLowerChar {
			charSetBuff.Write(lowercaseChars)
		}
		if argv.ContainUpperChar {
			charSetBuff.Write(uppercaseChars)
		}
		if argv.ContainSpecialChar {
			if argv.SpecialCharSet == "" {
				charSetBuff.Write(specialChars)
			} else {
				charSetBuff.WriteString(argv.SpecialCharSet)
			}
		}
		charSetLength := charSetBuff.Len()
		if charSetLength == 0 {
			return fmt.Errorf("charset is empty")
		}
		charSet := charSetBuff.Bytes()
		for i := 0; i < int(argv.Num); i++ {
			pw := make([]byte, argv.length)
			for j := 0; j < argv.length; j++ {
				pw[j] = charSet[rand.Intn(charSetLength)]
			}
			ctx.Write(pw)
			ctx.Write([]byte{'\n'})
		}
		return nil
	},
}

//--------------
// info command
//--------------
type infoT struct {
	cli.Helper
	Config
	All bool `cli:"a,all" usage:"show all found passwords"`
}

var info = &cli.Command{
	Name:        "info",
	Desc:        "show low-level information of password",
	Text:        "Usage: onepw info <ids...>",
	Argv:        func() interface{} { return new(infoT) },
	CanSubRoute: true,

	OnBefore: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*infoT)
		if argv.Help || len(ctx.Args()) == 0 {
			ctx.WriteUsage()
			return cli.ExitError
		}
		return nil
	},

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*infoT)
		return box.Inspect(ctx, ctx.Args(), argv.All)
	},
}
