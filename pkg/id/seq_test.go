package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	idSeq := New()
	require.NotNil(t, idSeq)
	assert.Equal(t, uint64(0), idSeq.seq)
}

func TestNext(t *testing.T) {
	idSeq := New()
	require.NotNil(t, idSeq)

	assert.Equal(t, uint64(1), idSeq.Next())
}

func TestRace(t *testing.T) {
	idSeq := New()
	require.NotNil(t, idSeq)

	for i := 0; i < 100; i++ {
		go assert.NotPanics(t, func() { idSeq.Next() })
	}
}

func TestConvertFromStringToArray(t *testing.T) {
	tcs := []struct {
		name        string
		in          string
		expectedOut []uint64
		expectedErr bool
	}{
		{
			name:        "empty",
			in:          "",
			expectedOut: nil,
			expectedErr: true,
		},
		{
			name:        "1",
			in:          "1",
			expectedOut: []uint64{1},
		},
		{
			name:        "1,2",
			in:          "1,2",
			expectedOut: []uint64{1, 2},
		},
	}

	for _, tc := range tcs {
		var (
			in          = tc.in
			expectedOut = tc.expectedOut
			expectedErr = tc.expectedErr
		)

		t.Run(tc.name, func(t *testing.T) {
			out, err := ConvertFromStringToArray(in)
			if expectedErr {
				require.Error(t, err)
			}

			assert.Equal(t, expectedOut, out)
		})
	}
}

func TestJoinIDArray(t *testing.T) {
	tcs := []struct {
		name        string
		inArr       []uint64
		delim       string
		expectedOut string
	}{
		{
			name:        "nil",
			inArr:       nil,
			delim:       ",",
			expectedOut: "",
		},
		{
			name:        "empty",
			inArr:       []uint64{},
			delim:       ",",
			expectedOut: "",
		},
		{
			name:        "normal",
			inArr:       []uint64{1, 2},
			delim:       ",",
			expectedOut: "1,2",
		},
	}

	for _, tc := range tcs {
		var (
			inArr       = tc.inArr
			delim       = tc.delim
			expectedOut = tc.expectedOut
		)
		t.Run(tc.name, func(t *testing.T) {
			out := JoinIDArray(inArr, delim)
			assert.Equal(t, expectedOut, out)
		})
	}
}
