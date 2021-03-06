package httputils

import (
	"net/http"
	"reflect"
	"testing"
)

type respRangeTest struct {
	r        string
	expRange *ContentRange
	expErr   bool
}

var respRangeTests = []respRangeTest{
	{r: "wrong", expErr: true},
	{r: "bytes", expErr: true},
	{r: "bytes ", expErr: true},
	{r: "1-2/11", expErr: true},
	{r: "bytes=1-2/11", expErr: true},
	{r: "bytes a-1/11", expErr: true},
	{r: "bytes 2-a/11", expErr: true},
	{r: "bytes -1-2/11", expErr: true},
	{r: "bytes 1--2/11", expErr: true},
	{r: "bytes 1a-2/11", expErr: true},
	{r: "bytes 1-2a/11", expErr: true},
	{r: "bytes 2-1/11", expErr: true},
	{r: "bytes 2-/11", expErr: true},
	{r: "bytes -1/11", expErr: true},
	{r: "bytes 10-11/11", expErr: true},
	{r: "bytes 0-0/0", expErr: true},
	{r: "bytes */11", expErr: true},
	{r: "bytes 1-2/11a", expErr: true},
	{r: "bytes 1-2/*", expErr: true},

	{r: "", expRange: nil},
	{r: "bytes 0-0/1", expRange: &ContentRange{0, 1, 1}},
	{r: "bytes 0-4/11", expRange: &ContentRange{0, 5, 11}},
	{r: "bytes 2-10/12", expRange: &ContentRange{2, 9, 12}},
	{r: "bytes 1-5/13", expRange: &ContentRange{1, 5, 13}},
	{r: "bytes 13-13/14", expRange: &ContentRange{13, 1, 14}},
}

func TestResponseContentRangeParsing(t *testing.T) {
	t.Parallel()
	for _, test := range respRangeTests {
		res, err := ParseResponseContentRange(test.r)

		if err != nil && !test.expErr {
			t.Errorf("Received an unexpected error for test %q: %s", test.r, err)
		}
		if err == nil && test.expErr {
			t.Errorf("Expected to receive an error for test %q", test.r)
		}
		if !reflect.DeepEqual(res, test.expRange) {
			t.Errorf("The received range for test %q '%#v' differ from the expected '%#v'", test.r, res, test.expRange)
		}
	}
}

func TestGetResponseRange(t *testing.T) {
	t.Parallel()
	headers := http.Header{"test": []string{"mest"}}

	if _, err := GetResponseRange(http.StatusOK, headers); err == nil {
		t.Error("Expected to receive an error for missing content-length header")
	}

	headers.Add("Content-Length", "123")
	expCr1 := &ContentRange{Start: 0, Length: 123, ObjSize: 123}
	if _, err := GetResponseRange(http.StatusAccepted, headers); err == nil {
		t.Error("Expected to receive an error for wrong status")
	}
	if cr, err := GetResponseRange(http.StatusOK, headers); err != nil {
		t.Errorf("Received an unexpected error: %s", err)
	} else if !reflect.DeepEqual(cr, expCr1) {
		t.Errorf("Expected %v to be equal to %v", cr, expCr1)
	}

	if _, err := GetResponseRange(http.StatusPartialContent, headers); err == nil {
		t.Error("Expected to receive an error for missing content-range header")
	}
	headers.Add("Content-Range", "bytes 3-22/99")
	expCr2 := &ContentRange{Start: 3, Length: 20, ObjSize: 99}
	if cr, err := GetResponseRange(http.StatusPartialContent, headers); err != nil {
		t.Errorf("Received an unexpected error: %s", err)
	} else if !reflect.DeepEqual(cr, expCr2) {
		t.Errorf("Expected %v to be equal to %v", cr, expCr2)
	}
}
