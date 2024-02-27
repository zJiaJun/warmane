package errors

import (
	"errors"
	"github.com/golang/glog"
)

var (
	ErrCsrfToken       = errors.New("查询获取csrfToken错误")
	ErrConfNotFound    = errors.New("配置文件[config.yml]未找到, 请把配置文件放到程序同一目录下")
	ErrConfDecodeError = errors.New("配置文件[config.yml]解析错误, 请检查配置文件")
)

func HandleError(err error) {
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, ErrConfNotFound):
		glog.Error(err.Error())
	case errors.Is(err, ErrConfDecodeError):
		glog.Error(err.Error())
	case errors.Is(err, ErrConfNotFound):
		glog.Error(err.Error())
	default:
		glog.Error(err.Error())
	}
}
