package throttle

type Throttle struct {
	waitChan chan struct{}
}

func NewThrottle(size int) (t Throttle) {
	t.waitChan = make(chan struct{}, size)
	return
}

func (t *Throttle) Done() {
	<-t.waitChan
}
func (t *Throttle) Wait() {
	t.waitChan <- struct{}{}
}
