package xunit

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"koality/util/find"
	"os"
)

type TestCase struct {
	ClassName string  `xml:"classname,attr" json:"-"`
	TestName  string  `xml:"name,attr" json:"-"`
	Path      string  `xml:"-" json:"path"`
	Name      string  `xml:"-" json:"name"`
	Time      float64 `xml:"time,attr" json:"time"`
	Failure   string  `xml:"failure" json:"failure"`
	Error     string  `xml:"error" json:"error"`
	Sysout    string  `xml:"system-out" json:"sysout"`
	Syserr    string  `xml:"system-err" json:"syserr"`
}

type TestSuite struct {
	XMLName   xml.Name    `xml:"testsuite" json:"-"`
	Name      string      `xml:"name,attr" json:"-"`
	Path      string      `xml:"-" json:"-"`
	TestCases []*TestCase `xml:"testcase" json:"testcases"`
}

func GetXunitResults(pattern string, paths []string, stdOut io.Writer, stdErr io.Writer) ([]byte, error) {
	var testSuites []*TestSuite
	userWalkFunc := func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		suite := &TestSuite{Path: path}
		if err = xml.Unmarshal(content, suite); err != nil {
			return err
		}

		testSuites = append(testSuites, suite)
		return nil
	}

	for _, path := range paths {
		if err := find.Find(path, pattern, userWalkFunc); err != nil {
			fmt.Fprintln(stdErr, err)
			return nil, err
		}
	}

	var testCases []*TestCase
	for _, testSuite := range testSuites {
		for _, testCase := range testSuite.TestCases {
			testCase.Path = testSuite.Path
			if len(testCase.ClassName) > 0 {
				testCase.Name = testCase.ClassName + "." + testCase.TestName
			} else {
				testCase.Name = testCase.TestName
			}
		}
		testCases = append(testCases, testSuite.TestCases...)
	}
	json, err := json.Marshal(testCases)
	if err != nil {
		fmt.Fprintln(stdErr, err)
		return nil, err
	}

	stdOut.Write(json)
	return json, nil
}
