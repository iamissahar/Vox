package lokisync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type LokiSyncer struct {
	url    string
	client *http.Client
}

func New(lokiURL, app string) *LokiSyncer {
	return &LokiSyncer{
		url:    lokiURL + "/loki/api/v1/push",
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type lokiStream struct {
	Streams []stream `json:"streams"`
}

type stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func closeReader(rd io.ReadCloser) {
	if rd != nil {
		_ = rd.Close()
	}
}

func (l *LokiSyncer) Write(p []byte) (int, error) {
	cp := make([]byte, len(p))
	copy(cp, p)
	payload := lokiStream{
		Streams: []stream{{
			Stream: map[string]string{"app": "vox"},
			Values: [][]string{
				{strconv.FormatInt(time.Now().UnixNano(), 10), string(cp)},
			},
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	resp, err := l.client.Post(l.url, "application/json", bytes.NewReader(body)) //nolint:bodyclose
	if err != nil {
		return 0, err
	}
	defer closeReader(resp.Body)
	if resp.StatusCode/100 != 2 {
		return 0, fmt.Errorf("loki returned %d", resp.StatusCode)
	}
	return len(p), nil
}

func (l *LokiSyncer) Sync() error { return nil }

// --- Buffered ---

type bufferedLokiSyncer struct {
	*LokiSyncer
	buf  chan []byte
	done chan struct{}
}

func NewBuffered(lokiURL, app string, bufSize int) *bufferedLokiSyncer {
	s := &bufferedLokiSyncer{
		LokiSyncer: New(lokiURL, app),
		buf:        make(chan []byte, bufSize),
		done:       make(chan struct{}),
	}
	go s.flush()
	return s
}

func (b *bufferedLokiSyncer) Write(p []byte) (int, error) {
	cp := make([]byte, len(p))
	copy(cp, p)
	select {
	case b.buf <- cp:
	default:
	}
	return len(p), nil
}

func (b *bufferedLokiSyncer) Sync() error {
	close(b.done)
	return nil
}

func (b *bufferedLokiSyncer) flush() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var batch [][]byte
	for {
		select {
		case p := <-b.buf:
			batch = append(batch, p)
			if len(batch) >= 100 {
				b.sendBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				b.sendBatch(batch)
				batch = nil
			}
		case <-b.done:
			for len(b.buf) > 0 {
				batch = append(batch, <-b.buf)
			}
			if len(batch) > 0 {
				b.sendBatch(batch)
			}
			return
		}
	}
}

func (b *bufferedLokiSyncer) sendBatch(batch [][]byte) {
	values := make([][]string, len(batch))
	for i, p := range batch {
		values[i] = []string{strconv.FormatInt(time.Now().UnixNano(), 10), string(p)}
	}
	payload := lokiStream{
		Streams: []stream{{
			Stream: map[string]string{"app": "vox"},
			Values: values,
		}},
	}
	body, _ := json.Marshal(payload)
	resp, err := b.client.Post(b.url, "application/json", bytes.NewReader(body))
	if err == nil {
		_ = resp.Body.Close()
	}
}
