package integration_tests

import "encoding/json"

// StringOrArray decodes a test.json field written either as one string or as a
// list of strings, and yields a []string in both cases.
//
// It lives in a plain file, not in run_test.go, because HTTPCheck in
// http_check.go is a non-test type that embeds it: a _test.go file is invisible
// to `go build`, so the package would not compile.
type StringOrArray []string

func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = []string{str}
		return nil
	}

	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}

	*s = arr
	return nil
}
