package main
import (
	"fmt"
	"os"
	"io"
	"time"
	"log"
	"strings"
	"runtime"
	"strconv"
	"crypto/sha256"
	"github.com/schollz/progressbar/v3"
	"github.com/fatih/color"
	"github.com/briandowns/spinner"
	flag "github.com/spf13/pflag"
	ac "github.com/JoaoDanielRufino/go-input-autocomplete"
)

var device string
var input string
var force bool
var disable_hash bool

func init() {
	flag.StringVarP(&device, "device", "d", "", "target device")
	flag.StringVarP(&input, "input", "i", "", "input file")
	flag.BoolVarP(&force, "force", "f", false, "override safety features")
	flag.BoolVarP(&disable_hash, "no-hash", "n", false, "disable hash verification")
	flag.Parse()
}

func GetPath() string {
	path, err := ac.Read("[ " + color.YellowString("i") + " ]  Please input your image file: ")
	if err != nil {
		log.Fatal(err)
	}
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "~/") {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		return homedir + path[1:]
	}
	if path == "" {
		fmt.Println("[", color.RedString("!"), "]  No image given, retrying.")
		path = GetPath()
	}
	return path
}

func GetDest() string {
	PrintAvail()
	dest, err := ac.Read("[ " + color.YellowString("i") + " ]  Please input destination: ")
	if err != nil {
		log.Fatal(err)
	}
	dest = strings.TrimSpace(dest)
	if strings.HasPrefix(dest, "~/") {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		return homedir + dest[1:]
	}
	if dest == "" {
		fmt.Println("[", color.RedString("!"), "]  No destination given, retrying.")
		dest = GetDest()
	}
	return dest
}

func WriteImage(image *os.File, target *os.File, size int64) (int64, error) {
	bar := progressbar.NewOptions(int(size),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(25),
		progressbar.OptionSetDescription("Writing image file..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	writer := io.MultiWriter(target, bar)
	written, err := io.Copy(writer, image)
	if err != nil {
		return 0, err
	}
	return written, err
}

func PrintAvail() {
	if runtime.GOOS == "linux" {
		block, _ := os.ReadDir("/sys/block")
		if len(block) == 0 {
			return
		}
		var targets []string
		for _, device := range block {
			if strings.HasPrefix(device.Name(), "sd") {
				targets = append(targets, device.Name())
			}
			if strings.HasPrefix(device.Name(), "nvme") {
				targets = append(targets, device.Name())
			}
			if strings.HasPrefix(device.Name(), "vd") {
				targets = append(targets, device.Name())
			}
		}
		fmt.Println("Available devices:")
		for _, target := range targets {
			sizefile, _ := os.Open("/sys/block/" + target + "/size")
			sizeread, _ := io.ReadAll(sizefile)
			_ = sizefile.Close()
			sizestring := strings.ReplaceAll(string(sizeread), "\n", "")
			size, _ := strconv.Atoi(sizestring)
			size = size * 512
			size = size / 1024 / 1024 / 1024

			fmt.Print("     * ", "/dev/" + target)
			if size > 0 {
				fmt.Print(" [", size, "GB]\n")
			} else {
				fmt.Println("")
			}
		}
	}
}

func main() {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	if input == "" {
		if len(flag.Args()) == 0 {
			input = GetPath()
		} else if len(flag.Args()) > 0 {
			input = flag.Args()[0]
		}
	}

	if device == "" {
		if len(flag.Args()) == 0 {
			device = GetDest()
		} else if len(flag.Args()) > 0 {
			if input == flag.Args()[0] && len(flag.Args()) > 1 {
				device = flag.Args()[1]
			}	else if input != flag.Args()[0] && len(flag.Args()) > 0 {
				device = flag.Args()[0]
			}
		}
	}

	s.Prefix = "[ "
	s.Suffix = " ]  Getting file details"
	s.Start()
	statinput, err := os.Stat(input)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Getting file details                     ")
		log.Fatal(err)
	}
	statdevice, err := os.Stat(device)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Getting file details                     ")
		log.Fatal(err)
	}
	image, err := os.Open(input)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Getting file details                     ")
		log.Fatal(err)
	}
	var inputsize int64
	var inputisblock bool
	if statinput.Size() != 0 {
		inputsize = statinput.Size()
		inputisblock = false
	} else {
		inputsize, err = image.Seek(0, io.SeekEnd)
		inputisblock = true
		_, _ = image.Seek(0, 0)
	}
	target, err := os.OpenFile(device, os.O_RDWR, 0660)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Getting file details                     ")
		log.Fatal(err)
	}
	var targetsize int64
	var targetisblock bool
	if statdevice.Size() != 0 {
		targetsize = statdevice.Size()
		targetisblock = false
	} else {
		targetsize, err = target.Seek(0, io.SeekEnd)
		targetisblock = true
		_, _ = target.Seek(0, 0)
	}
	prehash := sha256.New()
	if ! (force || disable_hash) {
		if err != nil {
			s.Stop()
			fmt.Println("\r[", color.RedString("✘"), "]  Getting file details                     ")
			log.Fatal(err)
		}

		_, err = io.Copy(prehash, image)
		_, _ = image.Seek(0, 0)
	}
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Getting file details                     ")
		log.Fatal(err)
	} else {
		s.Stop()
		fmt.Println("\r[", color.GreenString("✓"), "]  Getting file details                     ")
	}
	inputmb := fmt.Sprint("[", inputsize / 1024 / 1024, "MB]")
	devicemb := fmt.Sprint("[", targetsize / 1024 / 1024, "MB]")
	var inputblock string
	var targetblock string
	if inputisblock == true {
		inputblock = "[Blockdevice]"
	} else {
		inputblock = "[File]"
	}
	if targetisblock == true {
		targetblock = "[Blockdevice]"
	} else {
		targetblock = "[File]"
	}
	fmt.Println("[", color.BlueString("i"), "]  Input device/file: " + input, inputmb, inputblock)
	fmt.Println("[", color.BlueString("i"), "]  Output device/file: " + device, devicemb, targetblock)
	if force == false {
		if inputsize > targetsize {
			fmt.Println("[", color.RedString("w"), "]", color.RedString(" Warning:"), "Input file seems to be bigger than the destination!")
		}
		fmt.Print(color.HiWhiteString("Do you want to continue? [y/N]: "))
		var yesno string
		_, _ = fmt.Scanln(&yesno)
		yesno = strings.TrimSpace(yesno)
		if ! (yesno == "y" || yesno == "Y") {
			log.Fatal("aborted")
		}
	}
	written, err := WriteImage(image, target, inputsize)
	_, _ = target.Seek(0, 0)
	if err != nil {
		fmt.Println("\r[", color.RedString("✘"), "]  Writing image,", written, "bytes written                                                    ")
		log.Fatal(err)
	} else {
		fmt.Println("\r[", color.GreenString("✓"), "]  Writing image,", written, "bytes written                                                   ")
	}

	s.Prefix = "[ "
	s.Suffix = " ]  Syncing"
	s.Start()
	err = image.Sync()
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Syncing                    ")
		log.Fatal(err)
	}
	err = target.Sync()
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Syncing                    ")
		log.Fatal(err)
	} else {
		s.Stop()
		fmt.Println("\r[", color.GreenString("✓"), "]  Syncing                  ")
	}
	if ! (force || disable_hash) {
		s.Prefix = "[ "
		s.Suffix = " ]  Verifying"
		s.Start()
		posthash := sha256.New()
		_, err = io.CopyN(posthash, target, inputsize)
		presum := fmt.Sprintf("%x", prehash.Sum(nil))
		postsum := fmt.Sprintf("%x", posthash.Sum(nil))
		if err != nil || presum != postsum {
			s.Stop()
			fmt.Println("\r[", color.RedString("✘"), "]  Verifying                    ")
			log.Fatal(err)
		} else {
			s.Stop()
			fmt.Println("\r[", color.GreenString("✓"), "]  Verifying                  ")
		}
	}
	err = image.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = target.Close()
	if err != nil {
		log.Fatal(err)
	}
}
