package testdata

func UnconditionalForwardInput() {
	block(1)
	goto L1
	block(2)
L1:
	block(3)
}

func UnconditionalForwardExpected() {
	block(1)
	if !true {
		block(2)
	}
	block(3)
}

func UnconditionalBackwardInput() {
	block(1)
L1:
	block(2)
	goto L1
	block(3)
}

func UnconditionalBackwardExpected() {
	block(1)
	for {
		block(2)
		if !true {
			break
		}
	}
	block(3)
}
