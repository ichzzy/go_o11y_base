package slice_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ichzzy/go_o11y_base/internal/utils/slice"
)

func TestMap(t *testing.T) {
	type Source struct {
		ID   int
		Name string
	}
	type Destination struct {
		ID       int
		FullName string
	}

	sourceData := []Source{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}
	expected := []Destination{
		{ID: 1, FullName: "Alice"},
		{ID: 2, FullName: "Bob"},
	}

	result := slice.Map(sourceData, func(s Source) Destination {
		return Destination{ID: s.ID, FullName: s.Name}
	})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Map() = %v, want %v", result, expected)
	}
}

func TestMapE(t *testing.T) {
	type In struct {
		ID   int
		Name string
	}
	type Out struct {
		ID       int
		FullName string
	}
	type args struct {
		inputs []In
		fn     func(in In) (Out, error)
	}

	tests := []struct {
		name    string
		args    args
		want    []Out
		wantErr error
	}{
		{
			name: "success",
			args: args{
				inputs: []In{
					{ID: 1, Name: "Alice"},
					{ID: 2, Name: "Bob"},
				},
				fn: func(in In) (Out, error) {
					return Out{ID: in.ID, FullName: in.Name}, nil
				},
			},
			want: []Out{
				{ID: 1, FullName: "Alice"},
				{ID: 2, FullName: "Bob"},
			},
			wantErr: nil,
		},
		{
			name: "error",
			args: args{
				inputs: []In{
					{ID: 1, Name: "Alice"},
					{ID: 2, Name: ""},
				},
				fn: func(in In) (Out, error) {
					if in.Name == "" {
						return Out{}, errors.New("name is empty")
					}
					return Out{ID: in.ID, FullName: in.Name}, nil
				},
			},
			want: []Out{
				{ID: 1, FullName: "Alice"},
				{ID: 2, FullName: "Bob"},
			},
			wantErr: errors.New("name is empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := slice.MapE(tt.args.inputs, tt.args.fn)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("MapE() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("MapE() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
