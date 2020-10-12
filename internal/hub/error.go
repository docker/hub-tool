/*
   Copyright 2020 Docker Hub Tool authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package hub

import "fmt"

type authenticationError struct {
}

func (a authenticationError) Error() string {
	return "authentication error"
}

// IsAuthenticationError check if the error type is an authentication error
func IsAuthenticationError(err error) bool {
	_, ok := err.(*authenticationError)
	return ok
}

type invalidTokenError struct {
	token string
}

func (i invalidTokenError) Error() string {
	return fmt.Sprintf("invalid authentication token %q", i.token)
}

// IsInvalidTokenError check if the error type is an invalid token error
func IsInvalidTokenError(err error) bool {
	_, ok := err.(*invalidTokenError)
	return ok
}
