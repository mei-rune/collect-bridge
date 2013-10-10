package commons

/* Circular buffer object */
type CircularBuffer struct {
	start    int           /* index of oldest element              */
	count    int           /* the count of elements                */
	elements []interface{} /* vector of elements                   */
}

func NewCircularBuffer(elements []interface{}) *CircularBuffer {
	return &CircularBuffer{elements: elements}
}

func (self *CircularBuffer) Init(elements []interface{}) {
	self.elements = elements
	self.start = 0
	self.count = 0
}

/* clear all elements.*/
func (self *CircularBuffer) Clear() {
	self.start = 0
	self.count = 0
}

func (self *CircularBuffer) IsFull() bool {
	return self.count == len(self.elements)
}

/* return true while size is 0, otherwise return false */
func (self *CircularBuffer) IsEmpty() bool {
	return 0 == self.count
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking isFull(). */
func (self *CircularBuffer) Push(elem interface{}) {
	end := (self.start + self.count) % len(self.elements)
	self.elements[end] = elem
	if self.count == len(self.elements) {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
	} else {
		self.count++
	}
}

func (self *CircularBuffer) Get(idx int) interface{} {
	if self.IsEmpty() {
		return nil
	}

	current := (self.start + idx) % len(self.elements)
	return self.elements[current]
}

/* Read oldest element. App must ensure !isEmpty() first. */
func (self *CircularBuffer) Pop() interface{} {
	if self.IsEmpty() {
		return nil
	}

	elem := self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	self.count--
	return elem
}

func (self *CircularBuffer) First() interface{} {
	if self.IsEmpty() {
		return nil
	}

	return self.elements[self.start]
}

func (self *CircularBuffer) Last() interface{} {
	if self.IsEmpty() {
		return nil
	}

	end := (self.start + self.count - 1) % len(self.elements)
	return self.elements[end]
}

/* Read all elements.*/
func (self *CircularBuffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *CircularBuffer) All() []interface{} {
	if 0 == self.count {
		return []interface{}{}
	}

	res := make([]interface{}, 0, self.count)
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
