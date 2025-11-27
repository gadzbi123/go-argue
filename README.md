# AI Debate CLI (go-argue)

AI Debate CLI is a terminal user interface (TUI) for running live debates between two local LLMs served by [Ollama](https://ollama.com). You choose two models, enter a topic, and watch them argue back and forth in a styled, scrollable view. When the debate ends, the full conversation is automatically copied to your clipboard.

## Features

- Run debates between any two Ollama models (configurable via flags).
- Streaming responses rendered in a Bubble Tea / Lip Gloss TUI.
- Autoscroll toggle and clean separation of turns with timestamps.
- Debate history automatically yanked to the clipboard when finished.

## Requirements

- Go (version compatible with `go 1.24.9` or newer).
- [Ollama](https://ollama.com) running locally (default `http://localhost:11434`).
- The models you want to use must already be pulled in Ollama (for example `ollama pull phi3:mini`, `ollama pull gemma3:4b`).

## Installation

```bash
go install ./...
```

Or build a binary:

```bash
go build -o ai-debate-cli .
```

## Usage

From the project directory (or with the installed binary on your `PATH`):

```bash
./ai-debate-cli -model1 phi3:mini -model2 gemma3:4b
```

Then:

- Type a debate topic in the input field.
- Press `Enter` to start the debate.
- Press `a` to toggle autoscroll.
- Press `q` or `Ctrl+C` to stop.

When the debate stops, the full transcript (with model names and timestamps) is copied to your clipboard.

## Demo Video

Demo video: [video.mp4](video.mp4)
