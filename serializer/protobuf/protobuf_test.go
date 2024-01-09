// Copyright (c) nano Author and TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package protobuf

import (
	"flag"
	"testing"

	"github.com/colin1989/battery/errors"
	"github.com/colin1989/battery/helper"
	"github.com/colin1989/battery/proto"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update .golden files")

func TestNewSerializer(t *testing.T) {
	t.Parallel()
	serializer := NewSerializer()
	assert.NotNil(t, serializer)
}

func TestMarshal(t *testing.T) {
	var marshalTables = map[string]struct {
		raw interface{}
		err error
	}{
		"test_ok":            {&proto.Response{Data: []byte("data"), Error: &proto.Error{Msg: "error"}}, nil},
		"test_not_a_message": {"invalid", errors.ErrWrongValueType},
	}
	serializer := NewSerializer()

	for name, table := range marshalTables {
		t.Run(name, func(t *testing.T) {
			result, err := serializer.Marshal(table.raw)
			gp := helper.FixtureGoldenFileName(t, t.Name())

			if table.err == nil {
				assert.NoError(t, err)
				if *update {
					t.Log("updating golden file")
					helper.WriteFile(t, gp, result)
				}

				expected := helper.ReadFile(t, gp)
				assert.Equal(t, expected, result)
			} else {
				assert.Equal(t, table.err, err)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	gp := helper.FixtureGoldenFileName(t, "TestMarshal/test_ok")
	data := helper.ReadFile(t, gp)

	var dest proto.Response
	var unmarshalTables = map[string]struct {
		expected interface{}
		data     []byte
		dest     interface{}
		err      error
	}{
		"test_ok":           {&proto.Response{Data: []byte("data"), Error: &proto.Error{Msg: "error"}}, data, &dest, nil},
		"test_invalid_dest": {&proto.Response{Data: []byte(nil)}, data, "invalid", errors.ErrWrongValueType},
	}
	serializer := NewSerializer()

	for name, table := range unmarshalTables {
		t.Run(name, func(t *testing.T) {
			result := table.dest
			err := serializer.Unmarshal(table.data, result)
			assert.Equal(t, table.err, err)
			if table.err == nil {
				assert.Equal(t, table.expected, result)
			}
		})
	}
}
