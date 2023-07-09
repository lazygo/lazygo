package controller

import (
	"golang.org/x/crypto/bcrypt"

	request "github.com/lazygo/lazygo/examples/app/request/user"
	"github.com/lazygo/lazygo/examples/framework"
	cacheModel "github.com/lazygo/lazygo/examples/model/cache"
	dbModel "github.com/lazygo/lazygo/examples/model/db"
	"github.com/lazygo/lazygo/examples/utils/errors"
)

type UserController struct {
	Ctx framework.Context
}

// Login 登录
func (c *UserController) Login(req *request.LoginRequest) (*request.TokenResponse, error) {

	mdlUser := dbModel.NewUserModel()
	fields := []interface{}{"uid", "username", "password"}
	cond := map[string]interface{}{
		"username": req.Username,
	}
	user, n, err := mdlUser.FetchRow(fields, cond)
	if err != nil {
		c.Ctx.Logger().Warn("[msg: fetch user fail] [error: db error] [cond: %v] [err: %v]", cond, err)
		return nil, errors.ErrInternalServerError
	}
	if n == 0 {
		return nil, errors.ErrUserNotExists
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.ErrUserPasswordError
	}

	// 签发token
	token, err := cacheModel.NewAuthUserCache().Set(user.UID)
	if err != nil {
		c.Ctx.Logger().Warn("[msg: create token fail] [err: %v]", err)
		return nil, errors.ErrUserTokenError
	}

	resp := &request.TokenResponse{Token: token}
	return resp, nil
}

func (c *UserController) Logout(req *request.LogoutRequest) (*request.LogoutResponse, error) {
	err := cacheModel.NewAuthUserCache().Forget(req.Authorization)
	if err != nil {
		c.Ctx.Logger().Warn("[msg: delete token fail] [err: %v]", err)
	}

	resp := &request.LogoutResponse{}

	return resp, nil
}
