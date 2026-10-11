/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"context"
	"errors"
	"fmt"

	"github.com/kowabunga-cloud/common/klog"
	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

func NewUserRouter() sdk.Router {
	return sdk.NewUserAPIController(&UserService{}, sdk.WithUserAPIErrorHandler(ApiErrorHandler))
}

type UserService struct{}

var errInvalidCredentials = errors.New("invalid credentials")

func (s *UserService) Login(ctx context.Context, userCredentials sdk.UserCredentials) (sdk.ImplResponse, error) {
	klog.Debugf("Login attempt from %s ...", userCredentials.Email)

	if userCredentials.Email == "" || userCredentials.Password == "" {
		return HttpBadParams(nil)
	}

	// unknown user, unfinalized account creation or wrong password all
	// result in the very same response, not to disclose registered accounts
	u, err := FindUserByEmail(userCredentials.Email)
	if err == nil && !u.Enabled {
		err = fmt.Errorf("user %s has not finalized account creation", u.Name)
	}
	if err == nil {
		err = u.Verify(userCredentials.Password)
	}
	if err != nil {
		klog.Warningf("Login failure for %s: %v", userCredentials.Email, err)
		return HttpUnauthorized(errInvalidCredentials)
	}

	// let's go with JWT
	jwt, err := u.JwtSession()
	if err != nil {
		return HttpInternalError(err)
	}

	return HttpOK(sdk.UserCredentials{
		Email: userCredentials.Email,
		Jwt:   jwt,
	})
}

func (s *UserService) Logout(ctx context.Context) (sdk.ImplResponse, error) {
	if ctxGetAuthMethod(ctx) != HttpHeaderAuthorization {
		klog.Warningf("Attempting to logout from a non JWT-based session; pointless")
		return HttpNoContent() // not an error
	}

	userId := ctxGetUserId(ctx)
	klog.Debugf("Trying to logout user %s ...", userId)

	user, err := FindUserByID(userId)
	if err != nil {
		return HttpInternalError(err)
	}

	klog.Debugf("Invalidating JWTtoken for user %s", userId)
	user.JWT = ""

	err = user.Save()
	if err != nil {
		return HttpInternalError(err)
	}

	return HttpNoContent()
}

func (s *UserService) CreateUser(ctx context.Context, user sdk.User) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("user", user))

	// check for params
	if user.Name == "" || user.Email == "" {
		return HttpBadParams(nil)
	}

	// ensure user does not already exists
	_, err := FindUserByName(user.Name)
	if err == nil {
		return HttpConflict(err)
	}

	// create user
	u, err := NewUser(user.Name, user.Description, user.Email, user.Role, valueOr(user.Notifications, false))
	if err != nil {
		return HttpServerError(err)
	}

	payload := u.Model()
	LogHttpResponse(payload)
	return HttpCreated(payload)
}

func (s *UserService) DeleteUser(ctx context.Context, userId string) (sdk.ImplResponse, error) {
	// ensure user exists
	u, err := FindUserByID(userId)
	if err != nil {
		return HttpNotFound(err)
	}

	// remove user
	err = u.Delete()
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *UserService) ListUsers(ctx context.Context) (sdk.ImplResponse, error) {
	return HttpOK(FindResourceIDs(MongoCollectionUserName))
}

func (s *UserService) ReadUser(ctx context.Context, userId string) (sdk.ImplResponse, error) {
	u, err := FindUserByID(userId)
	if err != nil {
		return HttpNotFound(err)
	}

	payload := u.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}

func (s *UserService) SetUserPassword(ctx context.Context, userId string, password sdk.Password) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("userId", userId))

	// check for params
	if password.Value == "" {
		return HttpBadParams(nil)
	}

	// ensure user exists
	u, err := FindUserByID(userId)
	if err != nil {
		return HttpNotFound(err)
	}

	// update user password
	err = u.UpdatePassword(password.Value)
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *UserService) ResetPassword(ctx context.Context, email sdk.UserEmail) (sdk.ImplResponse, error) {
	if email.Email == "" {
		return HttpBadParams(nil)
	}

	// always succeed, not to disclose registered accounts
	u, err := FindUserByEmail(email.Email)
	if err != nil {
		klog.Warningf("Password reset request for unknown email %s", email.Email)
		return HttpOK(nil)
	}

	// reset user password
	err = u.ResetPasswordRequest()
	if err != nil {
		return HttpInternalError(err)
	}

	return HttpOK(nil)
}

func (s *UserService) ResetUserPassword(ctx context.Context, userId string) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("userId", userId))

	u, err := FindUserByID(userId)
	if err != nil {
		return HttpNotFound(err)
	}

	// reset user password
	err = u.ResetPasswordRequest()
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *UserService) SetUserApiToken(ctx context.Context, userId string, expire bool, expirationDate string) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("userId", userId), RA("expire", expire), RA("expirationDate", expirationDate))

	u, err := FindUserByID(userId)
	if err != nil {
		return HttpNotFound(err)
	}

	// check if user already has a registered token
	var t *Token

	tokenName := fmt.Sprintf("%s-api-key", u.Name)
	t, err = FindTokenByName(tokenName)
	if err != nil {
		// can't find any token, will create a new one
		t, err = NewUserToken(userId, tokenName, "", expire, expirationDate)
		if err != nil {
			return HttpServerError(err)
		}
	}

	// update token's expiration date, if any
	err = t.Update(tokenName, "", expire, expirationDate)
	if err != nil {
		return HttpServerError(err)
	}

	_, err = t.SetNewApiKey(true)
	if err != nil {
		return HttpServerError(err)
	}

	payload := t.Model()
	return HttpOK(payload)
}

func (s *UserService) UpdateUser(ctx context.Context, userId string, user sdk.UserUpdate) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("userId", userId), RA("user", user))

	// ensure user exists
	u, err := FindUserByID(userId)
	if err != nil {
		return HttpNotFound(err)
	}

	// update user
	err = u.Update(user.Name, user.Description, user.Email, user.Role, valueOr(user.Notifications, u.NotificationsEnabled))
	if err != nil {
		return HttpServerError(err)
	}

	payload := u.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}
