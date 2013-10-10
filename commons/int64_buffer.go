package commons

/* Circular buffer object */
type Int64Buffer struct {
	start    int     /* index of oldest element              */
	count    int     /* the count of elements                */
	elements []int64 /* vector of elements                   */
}

func NewInt64Buffer(elements []int64) *Int64Buffer {
	return &Int64Buffer{elements: elements}
}

func (self *Int64Buffer) Init(elements []int64) {
	self.elements = elements
	self.start = 0
	self.count = 0
}

/* clear all elements.*/
func (self *Int64Buffer) Clear() {
	self.start = 0
	self.count = 0
}

func (self *Int64Buffer) IsFull() bool {
	return self.count == len(self.elements)
}

/* return true while size is 0, otherwise return false */
func (self *Int64Buffer) IsEmpty() bool {
	return 0 == self.count
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking isFull(). */
func (self *Int64Buffer) Push(elem int64) {
	end := (self.start + self.count) % len(self.elements)
	self.elements[end] = elem
	if self.count == len(self.elements) {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
	} else {
		self.count++
	}
}

func (self *Int64Buffer) Get(idx int) int64 {
	if self.IsEmpty() {
		return 0
	}

	current := (self.start + idx) % len(self.elements)
	return self.elements[current]
}

/* Read oldest element. App must ensure !isEmpty() first. */
func (self *Int64Buffer) Pop() int64 {
	if self.IsEmpty() {
		return 0
	}

	elem := self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	self.count--
	return elem
}

func (self *Int64Buffer) First() int64 {
	if self.IsEmpty() {
		return 0
	}

	return self.elements[self.start]
}

func (self *Int64Buffer) Last() int64 {
	if self.IsEmpty() {
		return 0
	}

	end := (self.start + self.count - 1) % len(self.elements)
	return self.elements[end]
}

/* Read all elements.*/
func (self *Int64Buffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *Int64Buffer) All() []int64 {
	if 0 == self.count {
		return []int64{}
	}

	res := make([]int64, 0, self.count)
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
