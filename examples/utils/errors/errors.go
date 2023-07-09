package errors

import "github.com/lazygo/lazygo/server"

var (
	ErrUnauthorized        = server.NewHTTPError(200, 401, "认证失败")
	ErrBadRequest          = server.NewHTTPError(200, 400, "请求失败")
	ErrParamsError         = server.NewHTTPError(200, 400, "参数错误")
	ErrInternalServerError = server.NewHTTPError(200, 500, "内部服务错误")

	ErrDbError = server.NewHTTPError(200, 1006, "数据库错误")

	ErrUploadCosFail        = server.NewHTTPError(200, 10101, "文件上传服务异常")
	ErrInvalidImageFormat   = server.NewHTTPError(200, 10104, "图片格式错误")
	ErrPasswordError        = server.NewHTTPError(200, 10105, "密码长度不得小于6位")
	ErrPasswordTooLongError = server.NewHTTPError(200, 10106, "密码长度太长")
	ErrUsernameError        = server.NewHTTPError(200, 10106, "仅支持使用邮箱或手机号登录")

	ErrUserNotExists     = server.NewHTTPError(200, 10206, "用户不存在")
	ErrUserPasswordError = server.NewHTTPError(200, 10207, "用户名或密码错误")
	ErrUserTokenError    = server.NewHTTPError(200, 10208, "登录失败，请重试~")
)
