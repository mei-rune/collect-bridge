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

func (self *CircularBuffer) IsFull() bool {
	return self.count == len(self.elements)
}

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

/* Read all elements.*/
func (self *CircularBuffer) Size() int {
	return self.count
}

/* Read all elements.*/
func (self *CircularBuffer) All() []interface{} {
	if 0 == self.count {
		return nil
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

// int main(int argc, char **argv) {
//     CircularBuffer cb;
//     ElemType elem = {0};

//     int testBufferSize = 10; /* arbitrary size */
//     cbInit(&cb, testBufferSize);

//     /* Fill buffer with test elements 3 times */
//     for (elem.value = 0; elem.value < 3 * testBufferSize; ++ elem.value)
//         cbWrite(&cb, &elem);

//     /* Remove and print all elements */
//     while (!cbIsEmpty(&cb)) {
//         cbRead(&cb, &elem);
//         printf("%d\n", elem.value);
//     }

//     cbFree(&cb);
//     return 0;
// }
