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

func SiblingTwoLabelsInput() {
	block(1)
	if cond(1) {
		goto L1
	}
	block(2)
L1:
	block(3)
	if cond(2) {
		goto L2
	}
	block(4)
L2:
	block(5)
}

func SiblingTwoLabelsExpected() {
	block(1)
	if !cond(1) {
		block(2)
	}
	block(3)
	if !cond(2) {
		block(4)
	}
	block(5)
}

func SiblingTwoLabelsLoopsInput() {
	block(1)
L1:
	block(2)
	if cond(1) {
		goto L1
	}
	block(3)
L2:
	block(4)
	if cond(2) {
		goto L2
	}
	block(5)
}

func SiblingTwoLabelsLoopsExpected() {
	block(1)
	for {
		block(2)
		if !cond(1) {
			break
		}
	}
	block(3)
	for {
		block(4)
		if !cond(2) {
			break
		}
	}
	block(5)
}

func SiblingEmptyLoopInput() {
L1:
	if cond(1) {
		goto L1
	}
}

func SiblingEmptyLoopExpected() {
	for {
		if !cond(1) {
			break
		}
	}
}
