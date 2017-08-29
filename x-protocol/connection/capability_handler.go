package connection

type Capability_handler interface {
	String() string
	Is_support() bool
	Get()
	Set() bool
	commit()
}


