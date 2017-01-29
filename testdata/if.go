package testdata

func DirectIfInput() {
	block(1)
	if cond(1) {
		block(2)
		if cond(2) {
			goto L1
		}
		block(3)
	}
	block(4)
L1:
	block(5)
}

func DirectIfExpected() {
	gotoL1 := false
	block(1)
	if cond(1) {
		block(2)
		gotoL1 = cond(2)
		if !gotoL1 {
			block(3)
		}
	}
	if !gotoL1 {
		block(4)
	}
	block(5)
}

func DirectElseInput() {
	block(1)
	if cond(1) {
		block(2)
	} else {
		block(3)
		if cond(2) {
			goto L1
		}
		block(4)
	}
	block(5)
L1:
	block(6)
}

func DirectElseExpected() {
	gotoL1 := false
	block(1)
	if cond(1) {
		block(2)
	} else {
		block(3)
		gotoL1 = cond(2)
		if !gotoL1 {
			block(4)
		}
	}
	if !gotoL1 {
		block(5)
	}
	block(6)
}

func DirectElseIfInput() {
	block(1)
	if cond(1) {
		block(2)
	} else if cond(2) {
		block(3)
		if cond(3) {
			goto L1
		}
		block(4)
	}
	block(5)
L1:
	block(6)
}

func DirectElseIfExpected() {
	gotoL1 := false
	block(1)
	if cond(1) {
		block(2)
	} else if cond(2) {
		block(3)
		gotoL1 = cond(3)
		if !gotoL1 {
			block(4)
		}
	}
	if !gotoL1 {
		block(5)
	}
	block(6)
}
