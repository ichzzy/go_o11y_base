package slice

// []A -> []B
func Map[In any, Out any](inputs []In, fn func(in In) Out) []Out {
	outputs := make([]Out, len(inputs))
	for i, v := range inputs {
		outputs[i] = fn(v)
	}

	return outputs
}

// []A -> []B, return error
func MapE[In any, Out any](inputs []In, fn func(in In) (Out, error)) ([]Out, error) {
	outputs := make([]Out, len(inputs))
	for i, v := range inputs {
		res, err := fn(v)
		if err != nil {
			return nil, err
		}
		outputs[i] = res
	}

	return outputs, nil
}
