package e8asm

type Locator interface {
	Locate(lab string) (uint32, bool)
}
