package log

type Option struct {
	DirPath        string
	MaxFileSize    string
	RotateDuration string
	Level          string
	BackCount      uint32
	BackTime       string
}

func (o *Option) apply() {
	if o.BackTime == "" {
		o.BackTime = "7d"
	}

	if o.Level == "" {
		o.Level = "info"
	}

	if o.MaxFileSize == "" {
		o.MaxFileSize = "500M"
	}

	if o.RotateDuration == "" {
		o.RotateDuration = "1h"
	}

	if o.DirPath == "" {
		o.DirPath = "./log/"
	}
	return
}
