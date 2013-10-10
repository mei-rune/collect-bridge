package sampling

type SnmpTestResult struct {
	Result  bool  `json:"result"`
	SendAt  int64 `json:"send_at"`
	RecvAt  int64 `json:"recv_at"`
	Elapsed int64 `json:"elapsed"` // us
}

/* Circular buffer object */
type snmpTestResultBuffer struct {
	start    int              /* index of oldest element              */
	count    int              /* the count of elements                */
	elements []SnmpTestResult /* vector of elements                   */
}

func NewSnmpTestResultBuffer(elements []SnmpTestResult) *snmpTestResultBuffer {
	return &snmpTestResultBuffer{elements: elements}
}

func (self *snmpTestResultBuffer) init(elements []SnmpTestResult) {
	self.elements = elements
	self.start = 0
	self.count = 0
}

func (self *snmpTestResultBuffer) IsFull() bool {
	return self.count == len(self.elements)
}

func (self *snmpTestResultBuffer) IsEmpty() bool {
	return 0 == self.count
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking isFull(). */
func (self *snmpTestResultBuffer) Push(elem SnmpTestResult) {
	end := (self.start + self.count) % len(self.elements)
	self.elements[end] = elem
	if self.count == len(self.elements) {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
	} else {
		self.count++
	}
}

/* Read oldest element. App must ensure !isEmpty() first. */
func (self *snmpTestResultBuffer) Pop() *SnmpTestResult {
	if self.IsEmpty() {
		return nil
	}

	elem := &self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	self.count--
	return elem
}

func (self *snmpTestResultBuffer) Get(idx int) *SnmpTestResult {
	if self.IsEmpty() {
		return nil
	}

	current := (self.start + idx) % len(self.elements)
	return &self.elements[current]
}

func (self *snmpTestResultBuffer) First() *SnmpTestResult {
	if self.IsEmpty() {
		return nil
	}
	return &self.elements[self.start]
}

func (self *snmpTestResultBuffer) Last() *SnmpTestResult {
	if self.IsEmpty() {
		return nil
	}

	end := (self.start + self.count - 1) % len(self.elements)
	return &self.elements[end]
}

/* Read all elements.*/
func (self *snmpTestResultBuffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *snmpTestResultBuffer) All() []SnmpTestResult {
	if 0 == self.count {
		return nil
	}

	res := make([]SnmpTestResult, 0, self.count)
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
func (self *snmpTestResultBuffer) Clear() {
	self.start = 0
	self.count = 0
}
