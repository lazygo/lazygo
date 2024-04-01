package errors

import "github.com/lazygo/lazygo/server"

var (
	ErrUnauthorized        = server.NewHTTPError(200, 401, "认证失败")
	ErrParamsError         = server.NewHTTPError(200, 400, "参数错误")
	ErrInternalServerError = server.NewHTTPError(200, 500, "内部服务错误")

	ErrDbError = server.NewHTTPError(200, 1006, "数据库错误")

	ErrUserNotExists     = server.NewHTTPError(200, 10001, "用户不存在")
	ErrUserExists        = server.NewHTTPError(200, 10002, "用户已存在")
	ErrUserPasswordError = server.NewHTTPError(200, 10003, "用户名或密码错误")
	ErrUserTokenError    = server.NewHTTPError(200, 10004, "登录失败，请重试~")
	ErrPasswordTooShort  = server.NewHTTPError(200, 10005, "密码长度不得小于6位")
	ErrPasswordTooLong   = server.NewHTTPError(200, 10006, "密码长度太长")
	ErrUsernameInvalid   = server.NewHTTPError(200, 10007, "仅支持使用邮箱或手机号登录")

	ErrUploadCosFail      = server.NewHTTPError(200, 10101, "文件上传服务异常")
	ErrInvalidImageFormat = server.NewHTTPError(200, 10110, "图片格式错误")
)
