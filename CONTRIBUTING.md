# Contributing

Thanks for looking. A few things that make a change easy to take:

1. Open an issue first for anything larger than a bug fix, so we agree on the shape.
2. Keep the change focused - one behaviour per pull request.
3. Run the checks before opening the PR:
- `go vet ./...`
- `go test ./...`
- `go build -o bin/contractfault ./cmd/contractfault`
- viewer: `npm --prefix viewer install --no-save typescript@5.8 && npx --prefix viewer tsc -p viewer/tsconfig.json --noEmit`
4. Add a test for the behaviour you changed.
5. Match the surrounding style; the codebase is deliberately small and flat.
