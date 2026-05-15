package types

import "net/http"

type Code int

const (
	CodeOK Code = 0
)

const (
	CodeInvalidArgument Code = 10001
	CodeUnauthorized    Code = 10002
	CodeForbidden       Code = 10003
	CodeNotFound        Code = 10004
	CodeConflict        Code = 10005
	CodeTooManyRequests Code = 10006
)

const (
	CodeInvalidCredential Code = 20001
	CodeUserDisabled      Code = 20002
)

const (
	CodeInternal   Code = 50001
	CodeDependency Code = 50002
)

type CodeCategory string

const (
	CategorySuccess  CodeCategory = "success"
	CategoryClient   CodeCategory = "client"
	CategoryBusiness CodeCategory = "business"
	CategorySystem   CodeCategory = "system"
)

func (c Code) Category() CodeCategory {
	switch {
	case c == CodeOK:
		return CategorySuccess
	case c >= 10000 && c < 20000:
		return CategoryClient
	case c >= 20000 && c < 30000:
		return CategoryBusiness
	case c >= 50000 && c < 60000:
		return CategorySystem
	default:
		return CategorySystem
	}
}

func (c Code) HTTPStatus() int {
	switch c {
	case CodeOK:
		return http.StatusOK
	case CodeInvalidArgument, CodeTooManyRequests:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeInvalidCredential, CodeUserDisabled:
		return http.StatusUnprocessableEntity
	default:
		if c.Category() == CategorySystem {
			return http.StatusInternalServerError
		}
		return http.StatusBadRequest
	}
}
