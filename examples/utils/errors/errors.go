package errors

import "github.com/lazygo/lazygo/server"

var (
	ErrInternalServerError = server.NewHTTPError(500, 500, "internal server error")

	ErrInvalidParams = server.NewHTTPError(200, 400, "invalid params")
	ErrUnauthorized  = server.NewHTTPError(200, 401, "unauthorized")

	ErrInvalidJsonData = server.NewHTTPError(200, 1004, "broken json data")
	ErrDBError         = server.NewHTTPError(200, 1006, "数据库错误")
	ErrSystemError     = server.NewHTTPError(200, 1007, "系统错误，请添加QQ群798312770反馈")

	ErrUserNotExists     = server.NewHTTPError(200, 10001, "用户不存在")
	ErrUserExists        = server.NewHTTPError(200, 10002, "用户已存在")
	ErrUserPasswordError = server.NewHTTPError(200, 10003, "用户名或密码错误")
	ErrUserTokenError    = server.NewHTTPError(200, 10004, "登录失败，请重试~")
	ErrPasswordTooShort  = server.NewHTTPError(200, 10005, "密码长度不得小于6位")
	ErrPasswordTooLong   = server.NewHTTPError(200, 10006, "密码长度太长")
	ErrUsernameInvalid   = server.NewHTTPError(200, 10007, "账号必须为手机号或邮箱地址")
	ErrInvalidMobile     = server.NewHTTPError(200, 10008, "手机号格式错误")
	ErrMobileAlreadyUsed = server.NewHTTPError(200, 10009, "手机号已被使用")
	ErrSendSmsFail       = func(text string) error { return server.NewHTTPError(200, 10010, "短信发送失败: "+text) }
	ErrSendEmailFail     = func(text string) error { return server.NewHTTPError(200, 10011, "邮件发送失败: "+text) }
	ErrVerifyCaptchaFail = server.NewHTTPError(200, 10010, "验证码错误")
	ErrNotTimeYet        = server.NewHTTPError(200, 10011, "not just yet")
	ErrTooManyAttempts   = server.NewHTTPError(200, 10012, "too many attempts")

	ErrUploadCosFail      = server.NewHTTPError(200, 10101, "文件上传服务异常")
	ErrInvalidImageFormat = server.NewHTTPError(200, 10110, "图片格式错误")
)
