package serialize

import (
	"errors"
)

type testStuffAd struct {
	F string
	A int
	B int64
	C []byte
	H float64
}

type testStuffAdData struct {
	X string
	Y int
	Z int64
	J []byte
	K float64
}

func (t *testStuffAd) NewDataInstance() Data {
	return &testStuffAdData{}
}

func (t *testStuffAd) Data() Data {
	return &testStuffAdData{t.F, t.A, t.B, t.C, t.H}
}

func (t *testStuffAd) SetData(a interface{}) error {
	ad, ok := a.(*testStuffAdData)
	if !ok {
		return errors.New("Wrong data")
	}

	t.F = ad.X
	t.A = ad.Y
	t.B = ad.Z
	t.C = ad.J
	t.H = ad.K

	return nil
}

func (ad *testStuffAdData) SerialTag() string {
	return ""
}

func (ad *testStuffAdData) Primitive() DataAdapter {
	t := &testStuffAd{}
	t.F = ad.X
	t.A = ad.Y
	t.B = ad.Z
	t.C = ad.J
	t.H = ad.K

	return t
}
