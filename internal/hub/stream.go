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

import (
	"errors"
)

// EOS represents the end of the stream
var EOS = errors.New("end of stream") //nolint:golint

// Stream represents a call that can have more results (like a paginated call)
type Stream interface {
	Recv() (interface{}, error)
	Send(interface{})
	Close()
}

type stream struct {
	c      chan interface{}
	closed bool
}

// NewStream returns a new stream
func NewStream() Stream {
	return &stream{
		c:      make(chan interface{}),
		closed: false,
	}
}

func (s *stream) Recv() (interface{}, error) {
	if s.closed {
		return nil, EOS
	}
	val, e := <-s.c
	if !e {
		return nil, EOS
	}
	return val, nil
}

func (s *stream) Send(val interface{}) {
	s.c <- val
}

func (s *stream) Close() {
	if !s.closed {
		close(s.c)
		s.closed = true
	}
}
