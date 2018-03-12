package resources

import (
	"fmt"
)

type Resources struct {
	vCPU int
	ram  int
}


func NewResources(vCPU int, RAM int) *Resources {
	return &Resources{vCPU, RAM}
}

func (r *Resources) CPU() int {
	return r.vCPU
}

func (r *Resources) RAM() int {
	return r.ram
}

func (r *Resources) SetCPU(cpu int) {
	r.vCPU = cpu
}

func (r *Resources) SetRAM(ram int) {
	r.ram = ram
}

func (r *Resources) Copy() *Resources {
	res := &Resources{}
	res.vCPU = r.vCPU
	res.ram = r.ram
	return res
}

func (r *Resources) ToString() string {
	return fmt.Sprintf("CPUs: %d RAM: %d", r.vCPU, r.ram)
}