package controller

import (
	"encoding/json"
	"fmt"
	"strconv"

	request "github.com/lazygo/lazygo/examples/app/request/user"
	"github.com/lazygo/lazygo/examples/framework"
	cacheModel "github.com/lazygo/lazygo/examples/model/cache"
	dbModel "github.com/lazygo/lazygo/examples/model/db"
	"github.com/lazygo/lazygo/examples/utils"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	Ctx framework.Context
}

// Register 注册
func (ctl *UserController) Register(req *request.RegisterRequest) (any, error) {

	// fetch user
	mdlUser := dbModel.NewUserModel()

	user := map[string]any{}
	if req.Type == utils.TypeEmail {
		user["email"] = req.Username
	}
	if req.Type == utils.TypeMobile {
		user["mobile"] = req.Username
	}
	ok, err := mdlUser.Exists(user)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: fetch user fail] [error: db error] [cond: %v] [err: %v]", user, err)
		return nil, errors.ErrInternalServerError
	}
	if ok {
		return nil, errors.ErrUserExists
	}

	// save user
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: params error] [err: %v]", err)
		return nil, errors.ErrParamsError
	}
	user["password"] = passwordHash
	uid, err := mdlUser.Insert(user)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: create user fail] [error: db error] [err: %v]", err)
		return nil, errors.ErrInternalServerError
	}

	// sign token
	token, err := cacheModel.NewAuthUserCache().Set(uint64(uid))
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: create token fail] [err: %v]", err)
		return nil, errors.ErrUserTokenError
	}

	resp := &request.TokenResponse{Token: token}
	return resp, nil
}

// Login 登录
func (ctl *UserController) Login(req *request.LoginRequest) (any, error) {

	_, code, err := ctl.login(req)
	if err != nil {
		return nil, err
	}

	resp := &request.TokenResponse{Token: code}
	return resp, nil
}

// login
// return: UserData, auth code, err
func (ctl *UserController) login(req *request.LoginRequest) (*dbModel.UserData, string, error) {

	// fetch user
	mdlUser := dbModel.NewUserModel()

	cond := map[string]any{}
	if req.Type == utils.TypeEmail {
		cond["email"] = req.Username
	}
	if req.Type == utils.TypeMobile {
		cond["mobile"] = req.Username
	}
	user, n, err := mdlUser.First(cond, "uid", "password")
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: fetch user fail] [error: db error] [cond: %v] [err: %v]", cond, err)
		return nil, "", errors.ErrInternalServerError
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
	code, err := cacheModel.NewAuthUserCache().Set(user.UID)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: create token fail] [err: %v]", err)
		return nil, "", errors.ErrUserTokenError
	}
	return user, code, nil
}

func (ctl *UserController) Profile(req *request.ProfileRequest) error {
	fmt.Println(ctl.Ctx.UID())
	fmt.Println(req.UID)
	mdlUser := dbModel.NewUserModel()
	user, n, err := mdlUser.FetchByUid(req.UID)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: fetch user fail] [error: db error] [uid: %d] [err: %v]", req.UID, err)
		return errors.ErrInternalServerError
	}
	if n == 0 {
		return errors.ErrUserExists
	}
	fmt.Println(user.Email.String)
	fmt.Println(user.Mobile.String)
	data, err := json.Marshal(user)
	if err != nil {
		ctl.Ctx.Logger().Error("[msg: marshal json fail] [user: %v] [err: %v]", user, err)
		return errors.ErrInternalServerError
	}
	return ctl.Ctx.HTMLBlob(200, append([]byte(strconv.FormatUint(ctl.Ctx.UID(), 10)+"\n"), data...))
}

func (ctl *UserController) Logout(req *request.LogoutRequest) (any, error) {
	err := cacheModel.NewAuthUserCache().Forget(req.Authorization)
	if err != nil {
		ctl.Ctx.Logger().Warn("[msg: delete token fail] [err: %v]", err)
	}

	resp := &request.LogoutResponse{}

	return resp, nil
}
