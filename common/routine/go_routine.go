package routine

func Run(name string, fn func()) {
	go fn()
}
