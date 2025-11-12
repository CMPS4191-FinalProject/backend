# Quote of the Day API

A simple Go API that returns a quote of the day at the `/quote` endpoint.

## How to Run

1. Make sure you have Go installed (https://golang.org/dl/)
2. Open a terminal in this project directory.
3. Run:

   go run main.go

4. Visit http://localhost:8080/quote in your browser or use curl:

   curl http://localhost:8080/quote

## API

- **GET /quote**
  - Returns a JSON object with a daily quote.

## Example Response

```
{
  "text": "The best way to get started is to quit talking and begin doing.",
  "author": "Walt Disney"
}
```
