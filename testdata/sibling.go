package testdata

func SiblingBeforeLabelInput() {
	block(1)
	if cond(1) {
		goto L1
	}
	block(2)
L1:
	block(3)
}

func SiblingBeforeLabelExpected() {
	block(1)
	if !cond(1) {
		block(2)
	}
	block(3)
}

func SiblingAfterLabelInput() {
	block(1)
L1:
	block(2)
	if cond(1) {
		goto L1
	}
	block(3)
}

func SiblingAfterLabelExpected() {
	block(1)
	for {
		block(2)
		if !cond(1) {
			break
		}
	}
	block(3)
}
