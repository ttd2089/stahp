package inject

type fooer interface {
	Foo()
}

type assignableToFooer struct{}

func (assignableToFooer) Foo() {}

type structWithUnexportedFields struct {
	id int
}
