package testdata

func RangeInput() {
	block(1)
	for k, v := range []int{0, 1} {
		block(2)
		if cond(1) {
			goto L1
		}
		block(3)
	}
	block(4)
L1:
	block(5)
}

func RangeExpected() {
	gotoL1 := false
	block(1)
	for k, v := range []int{0, 1} {
		block(2)
		gotoL1 = cond(1)
		if gotoL1 {
			break
		}
		block(3)
	}
	if !gotoL1 {
		block(4)
	}
	block(5)
}
