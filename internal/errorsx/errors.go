package errorsx

import (
	"fmt"

	"github.com/KL-Engineering/oauth2-server/internal/errorsx/category"
	"github.com/KL-Engineering/oauth2-server/internal/errorsx/code"
)

type Error struct {
	Category category.Category `json:"category"`
	Code     code.Code         `json:"code"`
	Param    string            `json:"param,omitempty"`
	Message  string            `json:"message,omitempty"`
}

type Errors struct {
	Errors []Error `json:"errors"`
}

type BadRequestOptions struct {
	Code    code.Code
	Message string
	Param   string
}

func BadRequestError(opts BadRequestOptions) Error {
	return Error{
		Category: category.INVALID_REQUEST,
		Code:     opts.Code,
		Message:  opts.Message,
	}
}

func InvalidArgumentError(argument string) Error {
	return BadRequestError(BadRequestOptions{
		Code:    code.INVALID_ARGUMENT,
		Message: fmt.Sprintf("'%s' not valid.", argument),
		Param:   argument,
	})
}

func RequiredHeaderError(param string) Error {
	return BadRequestError(BadRequestOptions{
		Code:    code.REQUIRED_HEADER,
		Message: fmt.Sprintf("Header '%s' is required.", param),
		Param:   param,
	})
}

func NotFoundError(resource string) Error {
	return Error{
		Category: category.NOT_FOUND,
		Code:     code.NOT_FOUND,
		Message:  fmt.Sprintf("Resource '%s' not found.", resource),
	}
}

func InternalError() Error {
	return Error{
		Category: category.INTERNAL,
		Code:     code.INTERNAL,
		Message:  "Internal server error.",
	}
}
