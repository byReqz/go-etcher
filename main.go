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

func ReadImage(path string) ([]byte, error) {
	readfile, err := os.Open(path)
	if err != nil {
		return make([]byte, 0), err
	}
	bar := progressbar.DefaultBytes(
  	  -1,
    	"Reading file",
	)
	content, err := io.ReadAll(io.TeeReader(readfile, bar))
	if err != nil {
		return make([]byte, 0), err
	}
	return content, err
}

func WriteImage(path string, destination string, size int64) (int64, error) {
	image, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	target, err := os.OpenFile(destination, os.O_RDWR, 0660)
	if err != nil {
		return 0, err
	}
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

func main() {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "[ "
	s.Suffix = " ]  Getting file details"
	// s.Start()	

	// if not given via arg
	path := GetPath()
	// if not given via arg
	dest := GetDest()
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
	written, err := WriteImage(path, dest, stat.Size())
	if err != nil {
		s.Stop()
		fmt.Println("\r[", color.RedString("✘"), "]  Writing image,", written, "bytes written")
		log.Fatal(err)
	} else {
		s.Stop()
		fmt.Println("\r[", color.GreenString("✓"), "]  Writing image,", written, "bytes written")
	}
}
