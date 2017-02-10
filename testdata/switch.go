package testdata

func SwitchJumpInsideCaseInput() {
	switch x := 0; x {
	case 0:
		block(1)
	L1:
		block(2)
		if cond(1) {
			goto L1
		}
	}
}

func SwitchJumpInsideCaseExpected() {
	switch x := 0; x {
	case 0:
		block(1)
		for {
			block(2)
			if !cond(1) {
				break
			}
		}
	}
}

func SwitchJumpOutsideCaseInput() {
	block(1)
L1:
	block(2)
	switch x := 0; x {
	case 0:
		block(3)
		if cond(1) {
			goto L1
		}
		block(4)
	}
	block(5)
}

func SwitchJumpOutsideCaseExpected() {
	gotoL1 := false
	block(1)
	for {
		block(2)
		switch x := 0; x {
		case 0:
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
