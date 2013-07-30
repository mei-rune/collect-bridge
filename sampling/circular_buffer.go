package sampling

/* Circular buffer object */
type circularBuffer struct {
	//int         size;   /* maximum number of elements           */
	start    int           /* index of oldest element              */
	end      int           /* index at which to write new element  */
	elements []interface{} /* vector of elements                   */
}

func newCircularBuffer(elements []interface{}) *circularBuffer {
	return &circularBuffer{elements: elements}
}

func (self *circularBuffer) isFull() bool {
	return (self.end+1)%len(self.elements) == self.start
}

func (self *circularBuffer) isEmpty() bool {
	return self.end == self.start
}

/* Write an element, overwriting oldest element if buffer is full. App can
   choose to avoid the overwrite by checking cbIsFull(). */
func (self *circularBuffer) push(elem interface{}) {
	self.elements[self.end] = elem
	self.end = (self.end + 1) % len(self.elements)
	if self.end == self.start {
		self.start = (self.start + 1) % len(self.elements) /* full, overwrite */
	}
}

/* Read oldest element. App must ensure !cbIsEmpty() first. */
func (self *circularBuffer) pop() interface{} {
	if self.isEmpty() {
		return nil
	}
	elem := self.elements[self.start]
	self.start = (self.start + 1) % len(self.elements)
	return elem
}

/* Read all elements.*/
func (self *circularBuffer) size() int {
	if self.end == self.start {
		return 0
	}

	if self.end > self.start {
		return self.end - self.start
	}

	count := len(self.elements) - self.start
	count += self.end
	return count
}

/* Read all elements.*/
func (self *circularBuffer) all() []interface{} {
	if self.end == self.start {
		return nil
	}

	if self.end > self.start {
		res := make([]interface{}, 0, self.end-self.start)
		for i := self.start; i < self.end; i++ {
			res = append(res, self.elements[i])
		}
		return res
	}

	res := make([]interface{}, 0, len(self.elements))
	for i := self.start; i < len(self.elements); i++ {
		res = append(res, self.elements[i])
	}
	for i := 0; i < self.end; i++ {
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
