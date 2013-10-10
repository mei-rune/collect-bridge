package poller

type samplingResult struct {
	sampled_at int64
	is_ok      bool
}

/* Circular buffer object */
type samplingBuffer struct {
	start    int              /* index of oldest element              */
	count    int              /* the count of elements                */
	elements []samplingResult /* vector of elements                   */
}

func newSamplingBuffer(elements []samplingResult) *samplingBuffer {
	return &samplingBuffer{elements: elements}
}

func (self *samplingBuffer) Init(elements []samplingResult) {
	self.elements = elements
	self.start = 0
	self.count = 0
}

/* clear all elements.*/
func (self *samplingBuffer) Clear() {
	self.start = 0
	self.count = 0
}

func (self *samplingBuffer) IsFull() bool {
	return self.count == len(self.elements)
}

/* return true while size is 0, otherwise return false */
func (self *samplingBuffer) IsEmpty() bool {
	return 0 == self.count
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking isFull(). */
func (self *samplingBuffer) Push(elem samplingResult) {
	end := (self.start + self.count) % len(self.elements)
	self.elements[end] = elem
	if self.count == len(self.elements) {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
	} else {
		self.count++
	}
}

func (self *samplingBuffer) Get(idx int) samplingResult {
	current := (self.start + idx) % len(self.elements)
	return self.elements[current]
}

/* Read oldest element. App must ensure !isEmpty() first. */
func (self *samplingBuffer) Pop() samplingResult {
	elem := self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	self.count--
	return elem
}

func (self *samplingBuffer) First() samplingResult {
	return self.elements[self.start]
}

func (self *samplingBuffer) Last() samplingResult {
	end := (self.start + self.count - 1) % len(self.elements)
	return self.elements[end]
}

/* Read all elements.*/
func (self *samplingBuffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *samplingBuffer) All() []samplingResult {
	if 0 == self.count {
		return []samplingResult{}
	}

	res := make([]samplingResult, 0, self.count)
	if self.count <= (len(self.elements) - self.start) {
		for i := self.start; i < (self.start + self.count); i++ {
			res = append(res, self.elements[i])
		}
		return res
	}

	for i := self.start; i < len(self.elements); i++ {
		res = append(res, self.elements[i])
	}
	for i := 0; len(res) < self.count; i++ {
		res = append(res, self.elements[i])
	}

	return res
}
