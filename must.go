package staticplug

func must0(err error) {
	if err != nil {
		panic(err)
	}
}

func must1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
