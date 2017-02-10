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
