package httpx

import (
	"encoding/json"
	"regexp"
	"strings"
)

func NewSSEItem(raw string) (s *SSEItem, err error) {
	s = &SSEItem{}
	s.Raw = raw
	err = s.parse()
	if err != nil {
		return
	}
	return
}

type SSEItem struct {
	Raw      string
	Data     string
	Metadata map[string]string
	Event    string
}

func (s *SSEItem) parse() (err error) {
	dataArr := strings.Split(s.Raw, "\n")
	for _, item := range dataArr {
		if len(item) == 0 {
			continue
		}

		{
			//查找event
			eventReg := regexp.MustCompile(`^event: (.*)`)
			matches := eventReg.FindStringSubmatch(string(item))
			if len(matches) > 0 {
				s.Event = matches[1]
				continue
			}
		}

		{
			//查找data
			eventReg := regexp.MustCompile(`^data: (.*)`)
			matches := eventReg.FindStringSubmatch(string(item))
			if len(matches) > 0 {
				s.Data = matches[1]
				continue
			}
		}

		{
			s.Metadata = make(map[string]string, len(dataArr))
			//查找其他
			eventReg := regexp.MustCompile(`^(.*): (.*)`)
			matches := eventReg.FindStringSubmatch(string(item))
			if len(matches) > 0 {
				s.Metadata[matches[1]] = matches[2]
				continue
			}
		}
	}
	return
}

func (s *SSEItem) ScanData(v any, unmarshal ...func(data []byte, v any) error) (err error) {
	if len(unmarshal) > 0 {
		return unmarshal[0]([]byte(s.Data), v)
	} else {
		return json.Unmarshal([]byte(s.Data), v)
	}
}
