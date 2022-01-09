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
)

func GetPath() string {
	fmt.Print("Path to image: ")
	var path string
	_, err := fmt.Scanln(&path)
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
	fmt.Print("Path to Destination: ")
	var dest string
	_, err := fmt.Scanln(&dest)
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
	bar := progressbar.DefaultBytes(
  	  size,
    	"Writing image",
	)
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

	// if not given via arg
	path := GetPath()
	// if not given via arg
	dest := GetDest()

	s.Prefix = "[ "
	s.Suffix = " ]  Getting file details"
	s.Start()
	stat, err := os.Stat(path)
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
	image, err := os.Open(path)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Opening files                   ")
		log.Fatal(err)
	}
	target, err := os.OpenFile(dest, os.O_RDWR, 0660)
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Opening files                   ")
		log.Fatal(err)
	} else {
		s.Stop()
		fmt.Println("\r[", color.GreenString("✓"), "]  Opening files                 ")
	}

	written, err := WriteImage(image, target, stat.Size())
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Writing image,", written, "bytes written")
		log.Fatal(err)
	} else {
		s.Stop()
		fmt.Println("\r[", color.GreenString("✓"), "]  Writing image,", written, "bytes written")
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
