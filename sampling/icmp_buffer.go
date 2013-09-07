package sampling

type IcmpResult struct {
	Result bool   `json:"result"`
	Ttl    uint64 `json:"ttl"`

	SampledAt int64 `json:"sampled_at"`
}

/* Circular buffer object */
type icmpBuffer struct {
	start    int          /* index of oldest element              */
	count    int          /* the count of elements                */
	elements []IcmpResult /* vector of elements                   */
}

func NewIcmpBuffer(elements []IcmpResult) *icmpBuffer {
	return &icmpBuffer{elements: elements}
}

func (self *icmpBuffer) init(elements []IcmpResult) {
	self.elements = elements
	self.start = 0
	self.count = 0
}

func (self *icmpBuffer) IsFull() bool {
	return self.count == len(self.elements)
}

func (self *icmpBuffer) IsEmpty() bool {
	return 0 == self.count
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking isFull(). */
func (self *icmpBuffer) BeginPush() *IcmpResult {
	end := (self.start + self.count) % len(self.elements)
	elem := &self.elements[end]
	if self.count == len(self.elements) {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
		self.count--
	}
	return elem
}

func (self *icmpBuffer) CommitPush() {
	if self.count != len(self.elements) {
		self.count++
	}
}

/* Read oldest element. App must ensure !isEmpty() first. */
func (self *icmpBuffer) Pop() *IcmpResult {
	if self.IsEmpty() {
		return nil
	}

	elem := &self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	self.count--
	return elem
}

func (self *icmpBuffer) Get(idx int) *IcmpResult {
	if self.IsEmpty() {
		return nil
	}

	current := (self.start + idx) % len(self.elements)
	return &self.elements[current]
}

func (self *icmpBuffer) First() *IcmpResult {
	if self.IsEmpty() {
		return nil
	}
	return &self.elements[self.start]
}

func (self *icmpBuffer) Last() *IcmpResult {
	if self.IsEmpty() {
		return nil
	}

	end := (self.start + self.count - 1) % len(self.elements)
	return &self.elements[end]
}

/* Read all elements.*/
func (self *icmpBuffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *icmpBuffer) All() []IcmpResult {
	if 0 == self.count {
		return nil
	}

	res := make([]IcmpResult, 0, self.count)
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
func (self *icmpBuffer) Clear() {
	self.start = 0
	self.count = 0
}
