package hub

import (
	"context"
	"errors"
	"io"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"go.uber.org/zap"
)

type Groq struct {
	ApiKey        string
	Model         string
	BaseURL       string
	transcription *StringChan
	tokens        *StringChan
	errors        *ErrorChan
	log           *zap.Logger
}

func (g *Groq) handleStream(ctx context.Context, stream *ssestream.Stream[openai.ChatCompletionChunk]) error {
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				select {
				case g.tokens.Ch <- content:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
	return nil
}

// https://api.groq.com/openai/v1
func (g *Groq) do(ctx context.Context) (err error) {
	g.log.Debug("Groq.handleStream", zap.Bool("ctx_is_nil", ctx == nil))
	client := openai.NewClient(
		option.WithAPIKey(g.ApiKey),
		option.WithBaseURL(g.BaseURL),
	)
	defer g.tokens.Close()

	for {
		select {
		case transcript, ok := <-g.transcription.Ch:
			if !ok {
				g.log.Debug("Groq transcription channel closed")
				return nil
			}

			stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
				Model: g.Model,
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.UserMessage(transcript),
				},
			})

			if err = g.handleStream(ctx, stream); err != nil {
				return err
			}

			if err = stream.Err(); err != nil && !errors.Is(err, io.EOF) {
				g.log.Error("Groq stream error", zap.Error(err))
				return err
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// for transcript := range g.transcription.Ch {
	// 	select {
	// 	case <-ctx.Done():
	// 		return ctx.Err()
	// 	default:
	// 	}
	// 	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
	// 		Model: g.Model,
	// 		Messages: []openai.ChatCompletionMessageParamUnion{
	// 			openai.UserMessage(transcript),
	// 		},
	// 	})

	// 	if err = g.handleStream(ctx, stream); err != nil {
	// 		return err
	// 	}

	// 	if err = stream.Err(); err != nil && err != io.EOF {
	// 		g.log.Error("Groq stream error", zap.Error(err))
	// 		return err
	// 	}
	// }
}
