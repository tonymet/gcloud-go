package throttle

type Throttle struct {
	waitChan chan bool
}

func NewThrottle(size int) (t Throttle) {
	t.waitChan = make(chan bool, size)
	return
}

func (t *Throttle) Done() {
	<-t.waitChan
}
func (t *Throttle) Wait() {
	t.waitChan <- true
}
