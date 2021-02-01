package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type fileItem struct {
	content        []byte
	lastModifyTime time.Time
}

type FileAsyncer struct {
	fileItems sync.Map
}

func NewFileAsyncer() *FileAsyncer {
	return &FileAsyncer{}
}

func (a *FileAsyncer) ContentType(file string) ContentType {
	if strings.HasSuffix(file, ".yml") {
		return T_YAML
	}

	return T_JSON
}

func (a *FileAsyncer) Get(file string) []byte {
	info, err := os.Stat(file)

	if err != nil {
		logger.Errorf("conf file[%s] not exist", file)
		return nil
	}

	item, ok := a.fileItems.Load(file)
	if ok && info.ModTime() == item.(fileItem).lastModifyTime {
		return item.(fileItem).content
	}

	logger.Debugf("reload conf file[%s]", file)

	content, err := ioutil.ReadFile(file)
	if err != nil {
		logger.Errorf("read conf file[%s] err:%v", file, err)
		return nil
	}

	a.fileItems.Store(file, fileItem{
		content:        content,
		lastModifyTime: info.ModTime(),
	})

	return content
}

func (a *FileAsyncer) Set(file string, content []byte) error {

	return fmt.Errorf("the method is not implement")
}

func (a *FileAsyncer) Watch(file string) chan struct{} {
	return nil
}
