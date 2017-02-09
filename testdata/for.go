package testdata

func ForForwardsInput() {
	block(1)
	for i := 1; i < 2; i++ {
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

func ForForwardsExpected() {
	gotoL1 := false
	block(1)
	for i := 1; i < 2; i++ {
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

func ForBackwardsInput() {
	block(1)
L1:
	block(2)
	for i := 1; i < 2; i++ {
		block(3)
		if cond(1) {
			goto L1
		}
		block(4)
	}
	block(5)
}

func ForBackwardsExpected() {
	gotoL1 := false
	block(1)
	for {
		block(2)
		for i := 1; i < 2; i++ {
			block(3)
			gotoL1 = cond(1)
			if gotoL1 {
				break
			}
			block(4)
		}
		if !gotoL1 {
			break
		}
	}
	block(5)
}
