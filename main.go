package main
import (
	"fmt"
	"os"
	"io"
	"time"
	"log"
	"strings"
	"github.com/schollz/progressbar/v3"
	"github.com/fatih/color"
	"github.com/briandowns/spinner"
	flag "github.com/spf13/pflag"
	ac "github.com/JoaoDanielRufino/go-input-autocomplete"
)

var device string
var input string
var force bool

func init() {
	flag.StringVarP(&device, "device", "d", "", "target device")
	flag.StringVarP(&input, "input", "i", "", "input file")
	flag.BoolVarP(&force, "force", "f", false, "override safety features")
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
	return path
}

func GetDest() string {
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

func Sync(image *os.File, target *os.File) error {
	err := image.Sync()
	if err != nil {
		return err
	}
	err = target.Sync()
	if err != nil {
		return err
	}
	err = image.Close()
	if err != nil {
		return err
	}
	err = target.Close()
	if err != nil {
		return err
	}
	return nil
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
	stat, err := os.Stat(input)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Getting file details                     ")
		log.Fatal(err)
	} else {
		s.Stop()
		fmt.Println("\r[", color.GreenString("✓"), "]  Getting file details                     ")
	}

	s.Prefix = "[ "
	s.Suffix = " ]  Opening files"
	s.Start()
	image, err := os.Open(input)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Opening files                   ")
		log.Fatal(err)
	}
	target, err := os.OpenFile(device, os.O_RDWR, 0660)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Opening files                   ")
		log.Fatal(err)
	} else {
		s.Stop()
		fmt.Println("\r[", color.GreenString("✓"), "]  Opening files                 ")
	}

	fmt.Println("[", color.BlueString("i"), "]  Input device/file:", input)
	fmt.Println("[", color.BlueString("i"), "]  Output device/file:", device)
	fmt.Print(color.HiWhiteString("Do you want to continue? [y/N]: "))
	var yesno string
	_, _ = fmt.Scanln(&yesno)
	yesno = strings.TrimSpace(yesno)
	if ! (yesno == "y" || yesno == "Y") {
		log.Fatal("aborted")
	}

	written, err := WriteImage(image, target, stat.Size())
	if err != nil {
		fmt.Println("\r[", color.RedString("✘"), "]  Writing image,", written, "bytes written                                                    ")
		log.Fatal(err)
	} else {
		fmt.Println("\r[", color.GreenString("✓"), "]  Writing image,", written, "bytes written                                                   ")
	}

	s.Prefix = "[ "
	s.Suffix = " ]  Syncing"
	s.Start()
	err = Sync(image, target)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Syncing                    ")
		log.Fatal(err)
	} else {
		s.Stop()
		fmt.Println("\r[", color.GreenString("✓"), "]  Syncing                  ")
	}
}
