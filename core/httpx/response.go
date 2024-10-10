package httpx

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goslacker/slacker/core/tool"
	"io"
	"net/http"
	"strings"
)

type Response struct {
	*http.Response
}

// Scan 从body中读取数据，并反序列化到v中
func (r *Response) Scan(v any, unmarshal ...func(data []byte, v any) error) (err error) {
	defer r.Body.Close()
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	if len(unmarshal) > 0 {
		return unmarshal[0](content, v)
	} else {
		return json.Unmarshal(content, v)
	}
}

func (r *Response) ScanSSE() (list []*SSEItem, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	itemCh, errCh := r.ScanSSEAsync(ctx)
	for item := range itemCh {
		list = append(list, item)
	}

	for err = range errCh {
	}

	return
}

func (r *Response) ScanSSEAsync(ctx context.Context) (sseCh chan *SSEItem, errCh chan error) {
	sseCh = make(chan *SSEItem)
	errCh = make(chan error, 1)
	scanner := bufio.NewScanner(r.Body)
	scanner.Split(tool.SplitByEmptyLine)
	go func() {
		defer func() {
			r.Body.Close()
			close(sseCh)
			close(errCh)
		}()
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
				sseItem, err := NewSSEItem(scanner.Text())
				if err != nil {
					errCh <- err
					return
				}
				if strings.ToLower(sseItem.Event) == "error" {
					errCh <- errors.New(string(sseItem.Data))
					return
				}
				if strings.ToLower(sseItem.Event) == "done" || strings.ToLower(sseItem.Event) == "close" {
					return
				}
				if sseItem.Event == "" {
					errCh <- fmt.Errorf("no event received: <%#v>", sseItem)
					return
				}
				sseCh <- sseItem
			}
		}
		err := scanner.Err()
		if err != nil {
			errCh <- err
		}
	}()

	return
}
