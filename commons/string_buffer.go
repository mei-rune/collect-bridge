package commons

/* Circular buffer object */
type StringBuffer struct {
	start    int      /* index of oldest element              */
	count    int      /* the count of elements                */
	elements []string /* vector of elements                   */
}

func NewStringBuffer(elements []string) *StringBuffer {
	return &StringBuffer{elements: elements}
}

func (self *StringBuffer) Init(elements []string) {
	self.elements = elements
	self.start = 0
	self.count = 0
}

/* clear all elements.*/
func (self *StringBuffer) Clear() {
	self.start = 0
	self.count = 0
}

func (self *StringBuffer) IsFull() bool {
	return self.count == len(self.elements)
}

/* return true while size is 0, otherwise return false */
func (self *StringBuffer) IsEmpty() bool {
	return 0 == self.count
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking isFull(). */
func (self *StringBuffer) Push(elem string) {
	end := (self.start + self.count) % len(self.elements)
	self.elements[end] = elem
	if self.count == len(self.elements) {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
	} else {
		self.count++
	}
}

func (self *StringBuffer) Get(idx int) string {
	if self.IsEmpty() {
		return ""
	}

	current := (self.start + idx) % len(self.elements)
	return self.elements[current]
}

/* Read oldest element. App must ensure !isEmpty() first. */
func (self *StringBuffer) Pop() string {
	if self.IsEmpty() {
		return ""
	}

	elem := self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	self.count--
	return elem
}

func (self *StringBuffer) First() string {
	if self.IsEmpty() {
		return ""
	}

	return self.elements[self.start]
}

func (self *StringBuffer) Last() string {
	if self.IsEmpty() {
		return ""
	}

	end := (self.start + self.count - 1) % len(self.elements)
	return self.elements[end]
}

/* Read all elements.*/
func (self *StringBuffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *StringBuffer) All() []string {
	if 0 == self.count {
		return []string{}
	}

	res := make([]string, 0, self.count)
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
