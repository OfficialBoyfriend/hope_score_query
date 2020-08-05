package pkg

type ParseError struct {
	Msg string
	Err error
}

func (pe ParseError) Error() string {
	return pe.Msg
}

func (pe ParseError) Unwrap() error {
	return pe.Err
}

type ParseSuccess struct {
	Msg string
}

func (pe ParseSuccess) Error() string {
	return pe.Msg
}