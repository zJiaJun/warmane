package errors

import (
	"errors"
	"gitub.com/zJiajun/warmane/logger"
)

var (
	ErrConfNotFound      = errors.New("配置文件[config.yml]未找到, 请把配置文件放到程序同一目录下")
	ErrConfDecodeError   = errors.New("配置文件[config.yml]解析错误, 请检查配置文件")
	ErrCookieNotFound    = errors.New("cookies文件未找到,请把配置文件放到程序同一目录下")
	ErrCookieNonMatchKey = errors.New("cookies文件存在,不匹配cookiesKey")
)

func HandleError(err error) {
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, ErrConfNotFound):
		logger.Error(err.Error())
	case errors.Is(err, ErrConfDecodeError):
		logger.Error(err.Error())
	case errors.Is(err, ErrConfNotFound):
		logger.Error(err.Error())
	default:
		logger.Error(err.Error())
	}
}
