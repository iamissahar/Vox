package hub

import (
	"context"
	"errors"
	"io"

	fishaudio "github.com/fishaudio/fish-audio-go"
	"go.uber.org/zap"
)

type FishClient interface {
	StreamWebSocket(ctx context.Context, textChan <-chan string, params *fishaudio.StreamParams, opts *fishaudio.WebSocketOptions) (*fishaudio.WebSocketAudioStream, error)
}

type FishStream interface {
	Bytes() []byte
	Close() error
	Collect() ([]byte, error)
	Err() error
	Next() bool
	Read(p []byte) (n int, err error)
}

type FishAudio interface {
	StreamWebSocket(ctx context.Context, textChan <-chan string, params *fishaudio.StreamParams, opts *fishaudio.WebSocketOptions) (FishStream, error)
	HandleStream(stream FishStream)
	Do(ctx context.Context) error
}

type FishBuilder interface {
	SetReference(audio []byte, text string)
	SetHub(hub *Hub)
	SetTokens(tokens *StringChan)
	SetLogger(log *zap.Logger)
	Get() FishAudio
}

type Ref struct {
	Audio []byte
	Text  string
}

type FishHolder struct {
	client    FishClient
	Reference *Ref
	hub       *Hub
	tokens    *StringChan
	log       *zap.Logger
}

type BuildHolder struct {
	client    FishClient
	Reference *Ref
	hub       *Hub
	tokens    *StringChan
	log       *zap.Logger
}

func (b *BuildHolder) SetReference(audio []byte, text string) {
	b.Reference = &Ref{
		Audio: audio,
		Text:  text,
	}
}

func (b *BuildHolder) SetHub(hub *Hub) {
	b.hub = hub
}

func (b *BuildHolder) SetTokens(tokens *StringChan) {
	b.tokens = tokens
}

func (b *BuildHolder) SetLogger(log *zap.Logger) {
	b.log = log
}

func (b *BuildHolder) Get() FishAudio {
	return &FishHolder{
		client:    b.client,
		Reference: b.Reference,
		hub:       b.hub,
		tokens:    b.tokens,
		log:       b.log,
	}
}

func (f *FishHolder) HandleStream(stream FishStream) {
	f.log.Debug("FishHolder.HandleStream started")
	for stream.Next() {
		chunk := stream.Bytes()
		f.hub.Publish(chunk)
	}
	f.log.Debug("FishHolder.HandleStream finished")
}

func (f *FishHolder) StreamWebSocket(ctx context.Context, textChan <-chan string, params *fishaudio.StreamParams, opts *fishaudio.WebSocketOptions) (FishStream, error) {
	return f.client.StreamWebSocket(ctx, textChan, params, opts)
}

func (f *FishHolder) Do(ctx context.Context) error {
	f.log.Debug("FishHolder.Do", zap.Bool("ctx_is_nil", ctx == nil))
	stream, err := f.client.StreamWebSocket(ctx, f.tokens.Ch, &fishaudio.StreamParams{
		Latency: fishaudio.LatencyBalanced,
		References: []fishaudio.ReferenceAudio{{
			Audio: f.Reference.Audio,
			Text:  f.Reference.Text,
		}},
	}, nil)
	if err != nil {
		f.log.Error("Failed to stream TTS", zap.Error(err))
		return err
	}

	f.HandleStream(stream)

	err = stream.Err()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
