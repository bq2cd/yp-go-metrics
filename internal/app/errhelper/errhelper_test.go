package errhelper

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	errOne   = errors.New("error 1")
	errTwo   = errors.New("error 2")
	errThree = errors.New("error 3")
	errFour  = errors.New("error 4")
	errFive  = errors.New("error 5")
	errSix   = errors.New("error 6")
	errSeven = errors.New("error 7")
)

func TestUnwrapJoined(t *testing.T) {
	type args struct {
		errInitial error
	}
	type want struct {
		got []error
	}
	type testcase struct {
		args args
		want want
	}
	tests := map[string]testcase{
		"single error is returned as is": {
			args: args{errInitial: errOne},
			want: want{got: []error{errOne}},
		},
		"single level unwrapping": {
			args: args{errInitial: fmt.Errorf("%w, %w, %w", errOne, errTwo, errThree)},
			want: want{got: []error{errOne, errTwo, errThree}},
		},
		"double level unwrapping": {
			args: args{errInitial: errors.Join(fmt.Errorf("%w: %w: %w", errOne, errTwo, errThree), errFour, errFive)},
			want: want{got: []error{errOne, errTwo, errThree, errFour, errFive}},
		},
		"multi level unwrapping": {
			args: args{errInitial: errors.Join(
				errors.Join(errSix, fmt.Errorf("%w: %w: %w", errOne, errTwo, errThree)),
				errors.Join(errSeven, fmt.Errorf("%w %w", errFour, errFive)),
				errors.Join(errSeven, errors.Join(errSix, errFive), errors.Join(errFive, errFour, errThree)),
			)},
			want: want{got: []error{
				errSix,
				errOne, errTwo, errThree,
				errSeven,
				errFour, errFive,
				errSeven,
				errSix, errFive,
				errFive, errFour, errThree,
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := UnwrapJoined(tt.args.errInitial)
			assert.Equal(t, tt.want.got, got)
		})
	}
}
