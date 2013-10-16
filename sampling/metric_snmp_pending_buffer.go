package sampling

/* Circular buffer object */
type snmpPendingBuffer struct {
	start    int             /* index of oldest element              */
	count    int             /* the count of elements                */
	elements []testingRequst /* vector of elements                   */
}

func newPendingBuffer(elements []testingRequst) *snmpPendingBuffer {
	return &snmpPendingBuffer{elements: elements}
}

func (self *snmpPendingBuffer) Init(elements []testingRequst) {
	self.elements = elements
	self.start = 0
	self.count = 0
}

/* clear all elements.*/
func (self *snmpPendingBuffer) Clear() {
	self.start = 0
	self.count = 0
}

func (self *snmpPendingBuffer) IsFull() bool {
	return self.count == len(self.elements)
}

/* return true while size is 0, otherwise return false */
func (self *snmpPendingBuffer) IsEmpty() bool {
	return 0 == self.count
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking isFull(). */
func (self *snmpPendingBuffer) Push(elem testingRequst) {
	end := (self.start + self.count) % len(self.elements)
	self.elements[end] = elem
	if self.count == len(self.elements) {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
	} else {
		self.count++
	}
}

func (self *snmpPendingBuffer) Get(idx int) testingRequst {
	if self.IsEmpty() {
		return testingRequst{}
	}

	current := (self.start + idx) % len(self.elements)
	return self.elements[current]
}

/* Read oldest element. App must ensure !isEmpty() first. */
func (self *snmpPendingBuffer) Pop() testingRequst {
	if self.IsEmpty() {
		return testingRequst{}
	}

	elem := self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	self.count--
	return elem
}

func (self *snmpPendingBuffer) RemoveFirst(c int) {
	if self.count <= c {
		self.start = 0
		self.count = 0
		return
	}

	self.count -= c
	self.start = (self.start + c) % len(self.elements)
}

func (self *snmpPendingBuffer) First() testingRequst {
	if self.IsEmpty() {
		return testingRequst{}
	}

	return self.elements[self.start]
}

func (self *snmpPendingBuffer) Last() testingRequst {
	if self.IsEmpty() {
		return testingRequst{}
	}

	end := (self.start + self.count - 1) % len(self.elements)
	return self.elements[end]
}

/* Read all elements.*/
func (self *snmpPendingBuffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *snmpPendingBuffer) All() []testingRequst {
	if 0 == self.count {
		return []testingRequst{}
	}

	res := make([]testingRequst, 0, self.count)
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
