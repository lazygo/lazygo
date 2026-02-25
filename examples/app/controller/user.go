package controller

import (
	"cmp"
	stdErrors "errors"
	"strconv"
	"time"

	request "github.com/lazygo/lazygo/examples/app/request/user"
	"github.com/lazygo/lazygo/examples/framework"
	cacheModel "github.com/lazygo/lazygo/examples/model/cache"
	dbModel "github.com/lazygo/lazygo/examples/model/db"
	mailTemplates "github.com/lazygo/lazygo/examples/pkg/mail-templates"
	"github.com/lazygo/lazygo/examples/utils"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/pkg/sms"
	"github.com/lazygo/pkg/token/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	Ctx framework.Context
}

// Regist 注册
func (ctl *UserController) Regist(req *request.RegisterRequest) (any, error) {

	ok, err := ctl.verifyCaptcha("register", req.Username, req.Captcha)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: register verfiy captcha fail] [error: db error] [err: %v]", err)
		return nil, errors.ErrDBError
	}
	if !ok {
		ctl.Ctx.Logger().Warn("DEBUG: register verfiy captcha fail, %s, %s", req.Username, req.Captcha)
		return nil, errors.ErrVerifyCaptchaFail
	}

	// fetch user
	mdlUser := dbModel.NewUserModel(ctl.Ctx)

	user := map[string]any{
		"appid": dbModel.AppidMain,
	}
	if req.Type == utils.TypeEmail {
		user["email"] = req.Username
	}
	if req.Type == utils.TypeMobile {
		user["mobile"] = req.Username
	}
	ok, err = mdlUser.Exists(user)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: fetch user fail] [error: db error] [cond: %v] [err: %v]", user, err)
		return nil, errors.ErrDBError
	}
	if ok {
		return nil, errors.ErrUserExists
	}

	// save user
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: params error] [err: %v]", err)
		return nil, errors.ErrInvalidParams
	}
	user["password"] = passwordHash

	trans := func(tx *mysql.Tx, uid uint64) error {
		return nil
	}
	uid, err := mdlUser.Create(user, trans)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: create user fail] [error: db error] [err: %v]", err)
		return nil, errors.ErrDBError
	}

	// sign token
	code, err := cacheModel.NewAuthUserCache(ctl.Ctx).Set(dbModel.AppidMain, uint64(uid))
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: create token fail] [err: %v]", err)
		return nil, errors.ErrUserTokenError
	}
	token, err := JwtToken("m", uint64(uid), code)
	if err != nil {
		ctl.Ctx.Logger().Warn("[msg: sign jwt token fail] [err: %v]", err)
		return nil, errors.ErrUserTokenError
	}

	// 注册记录
	auditInfo := map[string]any{
		"referer":  req.Referer,
		"origin":   req.Origin,
		"username": req.Username,
	}
	dbModel.NewAuditModel(ctl.Ctx).Log(uint64(uid), dbModel.AuditTypeRegister, auditInfo, "")

	resp := &request.TokenResponse{Token: token}
	return resp, nil
}

// Forget 忘记密码
func (ctl *UserController) Forget(req *request.ForgetRequest) (any, error) {

	ok, err := ctl.verifyCaptcha("forget", req.Username, req.Captcha)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: register verfiy captcha fail] [error: db error] [err: %v]", err)
		return nil, errors.ErrDBError
	}
	if !ok {
		ctl.Ctx.Logger().Warn("DEBUG: register verfiy captcha fail, %s, %s", req.Username, req.Captcha)
		return nil, errors.ErrVerifyCaptchaFail
	}

	// fetch user
	mdlUser := dbModel.NewUserModel(ctl.Ctx)

	cond := map[string]any{
		"appid": dbModel.AppidMain,
	}
	if req.Type == utils.TypeEmail {
		cond["email"] = req.Username
	}
	if req.Type == utils.TypeMobile {
		cond["mobile"] = req.Username
	}
	user, n, err := mdlUser.First(cond, "uid")
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: fetch user fail] [error: db error] [cond: %v] [err: %v]", user, err)
		return nil, errors.ErrDBError
	}
	if n == 0 {
		return nil, errors.ErrUserNotExists
	}

	// save user
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: params error] [err: %v]", err)
		return nil, errors.ErrInvalidParams
	}

	_, err = mdlUser.UpdateByUid(user.UID, map[string]any{"password": passwordHash})
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: forget and reset password fail] [error: db error] [err: %v]", err)
		return nil, errors.ErrDBError
	}

	// sign token
	code, err := cacheModel.NewAuthUserCache(ctl.Ctx).Set(dbModel.AppidMain, user.UID)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: create token fail] [err: %v]", err)
		return nil, errors.ErrUserTokenError
	}
	token, err := JwtToken("m", user.UID, code)
	if err != nil {
		ctl.Ctx.Logger().Warn("[msg: sign jwt token fail] [err: %v]", err)
		return nil, errors.ErrUserTokenError
	}

	dbModel.NewAuditModel(ctl.Ctx).Log(user.UID, dbModel.AuditTypeForgetPassword, user, "")

	resp := &request.TokenResponse{Token: token}
	return resp, nil
}

// Login 登录
func (ctl *UserController) Login(req *request.LoginRequest) (any, error) {

	user, code, err := ctl.login(req)
	if err != nil {
		return nil, err
	}
	token, err := JwtToken("m", user.UID, code)
	if err != nil {
		ctl.Ctx.Logger().Warn("[msg: sign jwt token fail] [err: %v]", err)
		return nil, errors.ErrUserTokenError
	}

	// 登录记录
	auditInfo := map[string]any{
		"referer":  req.Referer,
		"origin":   req.Origin,
		"username": req.Username,
	}
	dbModel.NewAuditModel(ctl.Ctx).Log(user.UID, dbModel.AuditTypeLogin, auditInfo, "")

	resp := &request.TokenResponse{Token: token}
	return resp, nil
}

// login
// return: UserData, auth code, err
func (ctl *UserController) login(req *request.LoginRequest) (*dbModel.UserData, string, error) {

	// fetch user
	mdlUser := dbModel.NewUserModel(ctl.Ctx)

	cond := map[string]any{}
	if req.Type == utils.TypeEmail {
		cond["email"] = req.Username
	}
	if req.Type == utils.TypeMobile {
		cond["mobile"] = req.Username
	}
	user, n, err := mdlUser.First(cond, "uid", "appid", "password")
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: fetch user fail] [error: db error] [cond: %v] [err: %v]", cond, err)
		return nil, "", errors.ErrDBError
	}
	if n == 0 {
		return nil, "", errors.ErrUserNotExists
	}

	// verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, "", errors.ErrUserPasswordError
	}

	// sign token
	code, err := cacheModel.NewAuthUserCache(ctl.Ctx).Set(user.Appid, user.UID)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: create token fail] [err: %v]", err)
		return nil, "", errors.ErrUserTokenError
	}
	return user, code, nil
}

func (ctl *UserController) Logout(req *request.LogoutRequest) (any, error) {
	err := cacheModel.NewAuthUserCache(ctl.Ctx).Forget(req.Authorization)
	if err != nil {
		ctl.Ctx.Logger().Warn("[msg: delete token fail] [err: %v]", err)
	}

	resp := &request.LogoutResponse{}

	return resp, nil
}

// CheckLogin 检查登录
func (ctl *UserController) CheckLogin(req *request.CheckLoginRequest) (any, error) {
	return nil, nil
}

func (ctl *UserController) Profile(req *request.ProfileRequest) (any, error) {

	mdl := dbModel.NewUserModel(ctl.Ctx)
	user, _, err := mdl.FetchByUid(req.UID)
	if err != nil {
		ctl.Ctx.Logger().Warn("[msg: get user:%d failed] [err: %v]", user, err)
		return nil, errors.ErrDBError
	}

	resp := &request.ProfileResponse{}
	resp.Email = utils.Mask(user.Email.String, 2, 3)
	resp.Mobile = utils.Mask(user.Mobile.String, 2, 3)
	resp.UserName = utils.Mask(cmp.Or(resp.Mobile, resp.Email), 2, 3)
	resp.UserID = utils.EncryptUID(req.UID)
	resp.RegisterTime = time.Unix(user.CTime, 0).Format(time.DateOnly)
	return resp, nil
}

func (ctl *UserController) BindMobile(req *request.BindMobileRequest) (any, error) {

	ok, err := ctl.verifyCaptcha("bind", req.Mobile, req.Captcha)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: bind mobile verfiy captcha fail] [error: db error] [err: %v]", err)
		return nil, errors.ErrDBError
	}
	if !ok {
		ctl.Ctx.Logger().Warn("DEBUG: bind mobile verfiy captcha fail, %s, %s", req.Mobile, req.Captcha)
		return nil, errors.ErrVerifyCaptchaFail
	}

	mdlUser := dbModel.NewUserModel(ctl.Ctx)
	// 检测手机号是否被使用
	cond := map[string]any{
		"mobile": req.Mobile,
	}
	user, n, err := mdlUser.First(cond)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: get user fail] [error: db error] [err: %v]", err)
		return nil, errors.ErrDBError
	}
	if n > 0 && user.UID != req.UID {
		return nil, errors.ErrMobileAlreadyUsed
	}

	data := map[string]any{
		"mobile": req.Mobile,
	}
	_, err = mdlUser.UpdateByUid(req.UID, data)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: bind mobile fail] [error: db error] [err: %v]", err)
		return nil, errors.ErrDBError
	}

	mdlAuditLog := dbModel.NewAuditModel(ctl.Ctx)
	// 绑定记录
	bindInfo := map[string]any{
		"old": user.Mobile.String,
		"new": req.Mobile,
	}
	mdlAuditLog.Log(req.UID, dbModel.AuditTypeBindMobile, bindInfo, "")

	return &request.BindMobileResponse{}, nil
}

func (ctl *UserController) SendCaptcha(req *request.CaptchaRequest) (any, error) {
	cacheCaptcha := cacheModel.NewCaptchaCache(ctl.Ctx)
	code, err := cacheCaptcha.Store(req.Username, req.Opname)
	if err != nil {
		return nil, err
	}
	auditInfo := make(map[string]any)
	auditInfo["type"] = req.Type
	auditInfo["username"] = req.Username
	switch req.Type {
	case utils.TypeMobile:
		res, err := sms.SmsTo(req.Username, cacheModel.SmsTpl[req.Opname], []string{code, "30"})
		if err != nil {
			ctl.Ctx.Logger().Error("[msg: send sms fail] [resp: %s] [err: %v]", res, err)
			return nil, errors.ErrSendSmsFail(err.Error())
		}
		auditInfo["res"] = res

	case utils.TypeEmail:
		err := mailTemplates.SendRegisterCode(req.Username, code, req.Opname)
		if err != nil {
			ctl.Ctx.Logger().Error("[msg: send email fail] [err: %v]", err)
			return nil, errors.ErrSendEmailFail(err.Error())
		}

	default:
		return nil, stdErrors.New("auth type unsupported")
	}

	if req.UID != 0 {
		dbModel.NewAuditModel(ctl.Ctx).Log(uint64(req.UID), dbModel.AuditTypeSendCaptcha, auditInfo, "")
	}
	return &request.CaptchaResponse{}, nil
}

func (ctl *UserController) verifyCaptcha(opname string, username string, code string) (bool, error) {
	cacheCaptcha := cacheModel.NewCaptchaCache(ctl.Ctx)
	verifyCode, err := cacheCaptcha.Load(username, opname)
	if err != nil {
		return false, err
	}
	ctl.Ctx.Logger().Debug("[msg: verify code] [code: %s] [verify: %s]", code, verifyCode)
	return code == verifyCode, nil
}

func (ctl *UserController) ThirdLogin(req *request.ThirdLoginRequest) (any, error) {

	resp := &request.ThirdLoginResponse{
		Token: "test",
		URL:   req.Vendor,
		Extra: req.Extra,
	}
	return resp, nil
}

func JwtToken(audience string, uid uint64, code string) (string, error) {
	jwtEncoder, err := jwt.NewJwtEncoder([]byte(utils.PrivateKey))
	if err != nil {
		return "", err
	}

	// 签发jwt
	return jwtEncoder.Encode(&jwt.Claims{
		ID:       code,
		Audience: []string{audience},
		Issuer:   "lazygo.dev",
		Subject:  strconv.FormatUint(uid, 10),
	})
}
