package etcher

type Blockdevice struct {
	Path string
	Size int
	Type string
	//ID   string
}

// CheckHash checks if the given file has been properly applied to the given blockdevice.
//func CheckHash(f os.File, b Blockdevice) error {}
