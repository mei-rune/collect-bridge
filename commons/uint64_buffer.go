package commons

/* Circular buffer object */
type Uint64Buffer struct {
	start    int      /* index of oldest element              */
	count    int      /* the count of elements                */
	elements []uint64 /* vector of elements                   */
}

func NewUint64Buffer(elements []uint64) *Uint64Buffer {
	return &Uint64Buffer{elements: elements}
}

func (self *Uint64Buffer) Init(elements []uint64) {
	self.elements = elements
	self.start = 0
	self.count = 0
}

/* clear all elements.*/
func (self *Uint64Buffer) Clear() {
	self.start = 0
	self.count = 0
}

func (self *Uint64Buffer) IsFull() bool {
	return self.count == len(self.elements)
}

/* return true while size is 0, otherwise return false */
func (self *Uint64Buffer) IsEmpty() bool {
	return 0 == self.count
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking isFull(). */
func (self *Uint64Buffer) Push(elem uint64) {
	end := (self.start + self.count) % len(self.elements)
	self.elements[end] = elem
	if self.count == len(self.elements) {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
	} else {
		self.count++
	}
}

func (self *Uint64Buffer) Get(idx int) uint64 {
	if self.IsEmpty() {
		return 0
	}

	current := (self.start + idx) % len(self.elements)
	return self.elements[current]
}

/* Read oldest element. App must ensure !isEmpty() first. */
func (self *Uint64Buffer) Pop() uint64 {
	if self.IsEmpty() {
		return 0
	}

	elem := self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	self.count--
	return elem
}

func (self *Uint64Buffer) First() uint64 {
	if self.IsEmpty() {
		return 0
	}

	return self.elements[self.start]
}

func (self *Uint64Buffer) Last() uint64 {
	if self.IsEmpty() {
		return 0
	}

	end := (self.start + self.count - 1) % len(self.elements)
	return self.elements[end]
}

/* Read all elements.*/
func (self *Uint64Buffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *Uint64Buffer) All() []uint64 {
	if 0 == self.count {
		return nil
	}

	res := make([]uint64, 0, self.count)
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
