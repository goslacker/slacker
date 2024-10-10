package tool

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestSplitByEmptyLine(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("hello\n\nworld"))
	scanner.Split(SplitByEmptyLine)
	for scanner.Scan() {
		println(scanner.Text())
		println("---")
	}
	fmt.Printf("%+v\n", scanner.Err())
}

func TestSplitByEmptyLine2(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(`id: 0
event: Error
data: {"error_code":720702017,"error_message":"发生了一些错误，请尝试再次操作。如果问题仍然存在，请联系我们的支持团队。"}

id: 0
event: Error
data: {"error_code":720702017,"error_message":"发生了一些错误，请尝试再次操作。如果问题仍然存在，请联系我们的支持团队。"}

`))
	scanner.Split(SplitByEmptyLine)
	for scanner.Scan() {
		println(scanner.Text())
		println("---")
	}
	//fmt.Printf("%+v\n", scanner.Err())
}
