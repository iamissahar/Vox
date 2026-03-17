package hub

import (
	"context"
	"errors"
	"io"

	msginterfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/listen/v1/websocket/interfaces"
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/listen"
	websocketv1 "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/listen/v1/websocket"
	"go.uber.org/zap"
)

type Deepgram struct {
	ApiKey        string
	BaseURL       string
	Options       interfaces.LiveTranscriptionOptions
	client        *websocketv1.WSCallback
	transcription *StringChan
	errors        *ErrorChan
	log           *zap.Logger
	ctx           context.Context
}

func (d *Deepgram) Open(or *msginterfaces.OpenResponse) error {
	d.log.Debug("Deepgram.Open", zap.Any("or", or))
	d.log.Debug("Deepgram connection opened")
	return nil
}

func (d *Deepgram) Message(mr *msginterfaces.MessageResponse) error {
	d.log.Debug("Deepgram.Message", zap.Any("mr", mr))
	if mr != nil {
		if len(mr.Channel.Alternatives) > 0 {
			transcript := mr.Channel.Alternatives[0].Transcript
			if transcript != "" {
				d.log.Debug("Deepgram transcript received", zap.String("transcript", transcript))
				select {
				case d.transcription.Ch <- transcript:
				case <-d.ctx.Done():
					return d.ctx.Err()
				}
			} else {
				d.log.Warn("Deepgram transcript is empty", zap.Any("mr", mr))
			}
		} else {
			d.log.Warn("Deepgram transcript holder is empty", zap.Any("mr", mr))
		}
	} else {
		d.log.Warn("Deepgram message is nil")
	}
	return nil
}

func (d *Deepgram) Error(er *msginterfaces.ErrorResponse) error {
	d.log.Debug("Deepgram.Error", zap.Any("er", er))
	var err error
	if er != nil {
		d.log.Error("Deepgram error",
			zap.String("dg_code", er.ErrCode),
			zap.String("dg_description", er.Description),
			zap.String("dg_msg", er.ErrMsg),
			zap.String("dg_type", er.Type),
			zap.String("dg_variant", er.Variant),
		)
		err = errors.New(er.ErrMsg)
	} else {
		d.log.Error("Deepgram error", zap.String("err", "Unknown error"))
		err = errors.New("unknown error")
	}
	d.errors.Ch <- err
	return err
}

func (d *Deepgram) Close(cr *msginterfaces.CloseResponse) error {
	d.log.Debug("Deepgram.Close", zap.Any("cr", cr))
	defer d.transcription.Close()
	defer d.errors.Close()
	return nil
}

func (d *Deepgram) Metadata(md *msginterfaces.MetadataResponse) error            { return nil }
func (d *Deepgram) SpeechStarted(ssr *msginterfaces.SpeechStartedResponse) error { return nil }
func (d *Deepgram) UtteranceEnd(ur *msginterfaces.UtteranceEndResponse) error    { return nil }
func (d *Deepgram) UnhandledEvent(byData []byte) error                           { return nil }

func (d *Deepgram) handleStream(rd io.Reader) error {
	d.log.Debug("Deepgram.handleStream", zap.Bool("rd_is_nil", rd == nil))
	err := d.client.Stream(rd)
	if !errors.Is(err, io.EOF) {
		d.log.Error("Deepgram stream ends with an error", zap.Error(err))
		return err
	}
	return nil
}

func (d *Deepgram) do(rd io.Reader) (err error) {
	d.log.Debug("Deepgram.do", zap.Bool("ctx_is_nil", d.ctx == nil), zap.Bool("rd_is_nil", rd == nil))

	d.client, err = client.NewWSUsingCallback(d.ctx, d.ApiKey, &interfaces.ClientOptions{Host: d.BaseURL}, &d.Options, d)
	if err != nil {
		d.log.Error("Deepgram connection error", zap.Error(err))
		return err
	}
	if !d.client.Connect() {
		d.log.Error("Deepgram connection timed out")
		return errors.New("deepgram connection timed out")
	}

	if err = d.handleStream(rd); err != nil {
		return err
	}

	select {
	case err = <-d.errors.Ch:
	default:
	}

	return err
}
