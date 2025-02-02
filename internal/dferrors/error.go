/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dferrors

import (
	"errors"
	"fmt"

	commonv1 "d7y.io/api/pkg/apis/common/v1"
)

// common and framework errors
var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrInvalidHeader   = errors.New("invalid Header")
	ErrDataNotFound    = errors.New("data not found")
	ErrEmptyValue      = errors.New("empty value")
	ErrConvertFailed   = errors.New("convert failed")
	ErrEndOfStream     = errors.New("end of stream")
	ErrNoCandidateNode = errors.New("no candidate server node")
)

func IsEndOfStream(err error) bool {
	return err == ErrEndOfStream
}

type DfError struct {
	Code    commonv1.Code
	Message string
}

func (s *DfError) Error() string {
	return fmt.Sprintf("[%d]%s", s.Code, s.Message)
}

func New(code commonv1.Code, msg string) *DfError {
	return &DfError{
		Code:    code,
		Message: msg,
	}
}

func Newf(code commonv1.Code, format string, a ...any) *DfError {
	return &DfError{
		Code:    code,
		Message: fmt.Sprintf(format, a...),
	}
}

func CheckError(err error, code commonv1.Code) bool {
	if err == nil {
		return false
	}

	e, ok := err.(*DfError)

	return ok && e.Code == code
}
