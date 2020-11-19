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

package credentials

import (
	"time"

	dockercredentials "github.com/docker/cli/cli/config/credentials"
	clitypes "github.com/docker/cli/cli/config/types"
	"gopkg.in/square/go-jose.v2/jwt"
)

const (
	hubToolKey             = "hub-tool"
	hubToolTokenKey        = "hub-tool-token"
	hubToolRefreshTokenKey = "hub-tool-refresh-token"
	expirationWindow       = 1 * time.Minute
)

// Store stores and retrieves user auth information
// form the keystore
type Store interface {
	GetAuth() (*Auth, error)
	Store(auth Auth) error
	Erase()
}

// Auth represents user authentication
type Auth struct {
	Username string
	Password string
	// Token is the 2FA token
	Token string
	// RefreshToken is used to refresh the token when
	// it expires
	RefreshToken string
}

// TokenExpired returns true if the token is malformed or is expired,
// true otherwise
func (a *Auth) TokenExpired() bool {
	parsedToken, err := jwt.ParseSigned(a.Token)
	if err != nil {
		return true
	}

	out := jwt.Claims{}
	if err := parsedToken.UnsafeClaimsWithoutVerification(&out); err != nil {
		return true
	}
	if err := out.ValidateWithLeeway(jwt.Expected{Time: time.Now().Add(expirationWindow)}, 0); err != nil {
		return true
	}

	return false
}

type store struct {
	s dockercredentials.Store
}

// NewStore creates a new credentials store
func NewStore(provider func(string) dockercredentials.Store) Store {
	return &store{
		s: provider(hubToolKey),
	}
}

func (s *store) GetAuth() (*Auth, error) {
	auth, err := s.s.Get(hubToolKey)
	if err != nil {
		return nil, err
	}
	token, err := s.s.Get(hubToolTokenKey)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.s.Get(hubToolRefreshTokenKey)
	if err != nil {
		return nil, err
	}

	return &Auth{
		Username:     auth.Username,
		Password:     auth.Password,
		Token:        token.IdentityToken,
		RefreshToken: refreshToken.IdentityToken,
	}, nil
}

func (s *store) Store(auth Auth) error {
	if err := s.s.Store(clitypes.AuthConfig{
		Username:      auth.Username,
		IdentityToken: auth.Token,
		ServerAddress: hubToolTokenKey,
	}); err != nil {
		return err
	}
	if err := s.s.Store((clitypes.AuthConfig{
		Username:      auth.Username,
		IdentityToken: auth.RefreshToken,
		ServerAddress: hubToolRefreshTokenKey,
	})); err != nil {
		return err
	}
	return s.s.Store(clitypes.AuthConfig{
		Username:      auth.Username,
		Password:      auth.Password,
		ServerAddress: hubToolKey,
	})
}

func (s *store) Erase() {
	_ = s.s.Erase(hubToolKey)
	_ = s.s.Erase(hubToolRefreshTokenKey)
	_ = s.s.Erase(hubToolTokenKey)
}
