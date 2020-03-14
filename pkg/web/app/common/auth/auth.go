package auth

import (
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
)

type Register struct {
	Email      string `json:"email"`
	Pwd        string `json:"pwd"`
	ConfirmPwd string `json:"confirmPwd"`
}

func (_ *Register) Path() string {
	return "/me/register"
}

func (a *Register) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Register) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ResendActivateLink struct {
	Email string `json:"email"`
}

func (_ *ResendActivateLink) Path() string {
	return "/me/resendActivateLink"
}

func (a *ResendActivateLink) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ResendActivateLink) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Activate struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (_ *Activate) Path() string {
	return "/me/activate"
}

func (a *Activate) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Activate) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ChangeEmail struct {
	NewEmail string `json:"newEmail"`
}

func (_ *ChangeEmail) Path() string {
	return "/me/changeEmail"
}

func (a *ChangeEmail) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ChangeEmail) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ResendChangeEmailLink struct{}

func (_ *ResendChangeEmailLink) Path() string {
	return "/me/resendChangeEmailLink"
}

func (a *ResendChangeEmailLink) Do(c *app.Client) error {
	return app.Call(c, a.Path(), nil, nil)
}

func (a *ResendChangeEmailLink) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ConfirmChangeEmail struct {
	Me   ID     `json:"me"`
	Code string `json:"code"`
}

func (_ *ConfirmChangeEmail) Path() string {
	return "/me/confirmChangeEmail"
}

func (a *ConfirmChangeEmail) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ConfirmChangeEmail) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type ResetPwd struct {
	Email string `json:"email"`
}

func (_ *ResetPwd) Path() string {
	return "/me/resetPwd"
}

func (a *ResetPwd) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *ResetPwd) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type SetPwd struct {
	CurrentPwd    string `json:"currentPwd"`
	NewPwd        string `json:"newPwd"`
	ConfirmNewPwd string `json:"confirmNewPwd"`
}

func (_ *SetPwd) Path() string {
	return "/me/setPwd"
}

func (a *SetPwd) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *SetPwd) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Delete struct {
	Pwd string `json:"pwd"`
}

func (_ *Delete) Path() string {
	return "/me/delete"
}

func (a *Delete) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Delete) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Login struct {
	Email string `json:"email"`
	Pwd   string `json:"pwd"`
}

type LoginRes struct {
	Me ID `json:"id"`
}

func (_ *Login) Path() string {
	return "/me/login"
}

func (a *Login) Do(c *app.Client) (*LoginRes, error) {
	res := &LoginRes{}
	err := app.Call(c, a.Path(), a, res)
	return res, err
}

func (a *Login) MustDo(c *app.Client) *LoginRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Logout struct{}

func (_ *Logout) Path() string {
	return "/me/logout"
}

func (a *Logout) Do(c *app.Client) error {
	return app.Call(c, a.Path(), nil, nil)
}

func (a *Logout) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type Get struct{}

type GetRes LoginRes

func (_ *Get) Path() string {
	return "/me/get"
}

func (a *Get) Do(c *app.Client) (*GetRes, error) {
	res := &GetRes{}
	err := app.Call(c, a.Path(), nil, res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) *GetRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}