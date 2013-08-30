package sampling

/* Circular buffer object */
type fluxBuffer struct {
	start    int    /* index of oldest element              */
	count    int    /* the count of elements                */
	elements []Flux /* vector of elements                   */
}

func NewFluxBuffer(elements []Flux) *fluxBuffer {
	return &fluxBuffer{elements: elements}
}

func (self *fluxBuffer) init(elements []Flux) {
	self.elements = elements
	self.start = 0
	self.count = 0
}

func (self *fluxBuffer) IsFull() bool {
	return self.count == len(self.elements)
}

func (self *fluxBuffer) IsEmpty() bool {
	return 0 == self.count
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking isFull(). */
func (self *fluxBuffer) BeginPush() *Flux {
	end := (self.start + self.count) % len(self.elements)
	elem := &self.elements[end]
	if self.count == len(self.elements) {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
		self.count--
	}
	return elem
}

func (self *fluxBuffer) CommitPush() {
	if self.count != len(self.elements) {
		self.count++
	}
}

/* Read oldest element. App must ensure !isEmpty() first. */
func (self *fluxBuffer) Pop() *Flux {
	if self.IsEmpty() {
		return nil
	}

	elem := &self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	self.count--
	return elem
}

func (self *fluxBuffer) Get(idx int) *Flux {
	if self.IsEmpty() {
		return nil
	}

	current := (self.start + idx) % len(self.elements)
	return &self.elements[current]
}

func (self *fluxBuffer) First() *Flux {
	if self.IsEmpty() {
		return nil
	}
	return &self.elements[self.start]
}

func (self *fluxBuffer) Last() *Flux {
	if self.IsEmpty() {
		return nil
	}

	end := (self.start + self.count - 1) % len(self.elements)
	return &self.elements[end]
}

/* Read all elements.*/
func (self *fluxBuffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *fluxBuffer) All() []Flux {
	if 0 == self.count {
		return nil
	}

	res := make([]Flux, 0, self.count)
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

/* clear all elements.*/
func (self *fluxBuffer) Clear() {
	self.start = 0
	self.count = 0
}
