// Copyright (C) 2015 Scaleway. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.md file.

package cli

import (
	"fmt"
	"strings"

	"github.com/scaleway/scaleway-cli/pkg/commands"
)

var cmdRun = &Command{
	Exec:        runRun,
	UsageLine:   "run [OPTIONS] IMAGE [COMMAND] [ARG...]",
	Description: "Run a command in a new server",
	Help:        "Run a command in a new server.",
	Examples: `
    $ scw run ubuntu-bionic
    $ scw run --commercial-type=DEV1-S ubuntu-bionic
    $ scw run --show-boot --commercial-type=DEV1-S ubuntu-bionic
    $ scw run --rm ubuntu-bionic
    $ scw run -a --rm ubuntu-bionic
    $ scw run --gateway=myotherserver ubuntu-bionic
    $ scw run ubuntu-bionic bash
    $ scw run --name=mydocker docker docker run moul/nyancat:armhf
    $ scw run --bootscript=3.2.34 --env="boot=live rescue_image=http://j.mp/scaleway-ubuntu-trusty-tarball" 50GB bash
    $ scw run --attach alpine
    $ scw run --detach alpine
    $ scw run --tmp-ssh-key alpine
    $ scw run --userdata="FOO=BAR FILE=@/tmp/file" alpine
`,
}

func init() {
	cmdRun.Flag.StringVar(&runCreateName, []string{"-name"}, "", "Assign a name")
	cmdRun.Flag.StringVar(&runCreateBootscript, []string{"-bootscript"}, "", "Assign a bootscript")
	cmdRun.Flag.StringVar(&runCreateEnv, []string{"e", "-env"}, "", "Provide metadata tags passed to initrd (i.e., boot=rescue INITRD_DEBUG=1)")
	cmdRun.Flag.StringVar(&runCreateVolume, []string{"v", "-volume"}, "", "Attach additional volume (i.e., 50G)")
	cmdRun.Flag.BoolVar(&runHelpFlag, []string{"h", "-help"}, false, "Print usage")
	cmdRun.Flag.Int64Var(&runTimeout, []string{"T", "-timeout"}, 0, "Set timeout value to seconds")
	cmdRun.Flag.StringVar(&runIPAddress, []string{"-ip-address"}, "", "Assign a reserved public IP, a 'dynamic' one or 'none' (default to 'none' if gateway specified, 'dynamic' otherwise)")
	cmdRun.Flag.BoolVar(&runAttachFlag, []string{"a", "-attach"}, false, "Attach to serial console")
	cmdRun.Flag.BoolVar(&runDetachFlag, []string{"d", "-detach"}, false, "Run server in background and print server ID")
	cmdRun.Flag.StringVar(&runGateway, []string{"g", "-gateway"}, "", "Use a SSH gateway")
	cmdRun.Flag.StringVar(&runUserdatas, []string{"u", "-userdata"}, "", "Start a server with userdata predefined")
	cmdRun.Flag.StringVar(&runCommercialType, []string{"-commercial-type"}, "DEV1-S", "Start a server with specific commercial-type DEV1-[S|M|L|XL], GP1-[XS|S|M|L|XL], RENDER-S")
	cmdRun.Flag.StringVar(&runBootType, []string{"-boot-type"}, "auto", "Choose between 'local' and 'bootscript' boot")
	cmdRun.Flag.StringVar(&runSSHUser, []string{"-user"}, "root", "Specify SSH User")
	cmdRun.Flag.BoolVar(&runAutoRemove, []string{"-rm"}, false, "Automatically remove the server when it exits")
	cmdRun.Flag.BoolVar(&runIPV6, []string{"-ipv6"}, false, "Enable IPV6")
	cmdRun.Flag.BoolVar(&runTmpSSHKey, []string{"-tmp-ssh-key"}, false, "Access your server without uploading your SSH key to your account")
	cmdRun.Flag.BoolVar(&runShowBoot, []string{"-show-boot"}, false, "Allows to show the boot")
	cmdRun.Flag.IntVar(&runSSHPort, []string{"p", "-port"}, 22, "Specify SSH port")
	cmdRun.Flag.BoolVar(&runNoTTY, []string{"-no-tty"}, false, "Disable pseudo-terminal allocation")
	// FIXME: handle start --timeout
}

// Flags
var runCreateName string       // --name flag
var runAutoRemove bool         // --rm flag
var runCreateBootscript string // --bootscript flag
var runCreateEnv string        // -e, --env flag
var runCreateVolume string     // -v, --volume flag
var runIPAddress string        // --ip-address flag
var runHelpFlag bool           // -h, --help flag
var runAttachFlag bool         // -a, --attach flag
var runDetachFlag bool         // -d, --detach flag
var runGateway string          // -g, --gateway flag
var runUserdatas string        // -u, --userdata flag
var runCommercialType string   // --commercial-type flag
var runBootType string         // --boot-type flag
var runTmpSSHKey bool          // --tmp-ssh-key flag
var runShowBoot bool           // --show-boot flag
var runIPV6 bool               // --ipv6 flag
var runTimeout int64           // --timeout flag
var runSetState string         // --set-state flag
var runSSHUser string          // --user flag
var runSSHPort int             // -p, --port flag
var runNoTTY bool              // -no-tty flag

func runRun(cmd *Command, rawArgs []string) error {
	if runHelpFlag {
		return cmd.PrintUsage()
	}
	if len(rawArgs) < 1 {
		return cmd.PrintShortUsage()
	}
	if runAttachFlag && len(rawArgs) > 1 {
		return fmt.Errorf("conflicting options: -a and COMMAND")
	}
	if runAttachFlag && runDetachFlag {
		return fmt.Errorf("conflicting options: -a and -d")
	}
	if runAttachFlag && runShowBoot {
		return fmt.Errorf("conflicting options: -a and --show-boot")
	}
	if runShowBoot && len(rawArgs) > 1 {
		return fmt.Errorf("conflicting options: --show-boot and COMMAND")
	}
	if runShowBoot && runDetachFlag {
		return fmt.Errorf("conflicting options: --show-boot and -d")
	}
	if runDetachFlag && len(rawArgs) > 1 {
		return fmt.Errorf("conflicting options: -d and COMMAND")
	}
	if runAutoRemove && runDetachFlag {
		return fmt.Errorf("conflicting options: --detach and --rm")
	}

	args := commands.RunArgs{
		Attach:         runAttachFlag,
		Bootscript:     runCreateBootscript,
		Command:        rawArgs[1:],
		Detach:         runDetachFlag,
		Gateway:        runGateway,
		Image:          rawArgs[0],
		Name:           runCreateName,
		AutoRemove:     runAutoRemove,
		TmpSSHKey:      runTmpSSHKey,
		ShowBoot:       runShowBoot,
		IP:             runIPAddress,
		Timeout:        runTimeout,
		Userdata:       runUserdatas,
		CommercialType: runCommercialType,
		State:          runSetState,
		IPV6:           runIPV6,
		SSHUser:        runSSHUser,
		SSHPort:        runSSHPort,
		BootType:       runBootType,
		NoTTY:          runNoTTY,
		// FIXME: Timeout
	}

	if len(runCreateEnv) > 0 {
		args.Tags = strings.Split(runCreateEnv, " ")
	}
	if len(runCreateVolume) > 0 {
		args.Volumes = strings.Split(runCreateVolume, " ")
	}
	ctx := cmd.GetContext(rawArgs)
	return commands.Run(ctx, args)
}
