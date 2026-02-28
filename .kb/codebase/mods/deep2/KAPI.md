# KAPI — HTTP API Client Plain Language Guide

---

## What Is KAPI?

KAPI is a tool for talking to web servers directly — without opening a browser. Think of it like sending a letter and receiving a reply, instead of visiting someone's house in person.

When you test a website, sometimes you need to:
- Check if the login API works before testing the login page
- Create test data via API before the browser test starts
- Verify that submitting a form actually saved data on the server

KAPI lets you do all of this.

## How Does It Compare?

| Tool | Language | What It Does |
|------|----------|-------------|
| **KAPI** | Go (kexas) | Send HTTP requests, check responses |
| **Playwright APIRequestContext** | JavaScript | Same thing |
| **RestSharp** | C# | Same thing |
| **RestAssured** | Java | Same thing |

## Example

```go
client := kapi.NewClient("https://api.example.com", kapi.WithBearerToken("my-token"))

// GET a user
resp, err := client.Get("/users/1")
// resp.StatusCode = 200
// resp.Body = `{"id":1,"name":"John"}`

// POST a new user
resp, err = client.Post("/users", map[string]string{"name": "Jane"})
// resp.StatusCode = 201
```

## Key Features

| Feature | What It Means |
|---------|---------------|
| **Base URL** | Set once, used for all requests — no repeating `https://api.example.com` |
| **Default headers** | Set auth token once, sent with every request automatically |
| **Per-request overrides** | Override headers/cookies for one specific request |
| **JSON helpers** | Send and receive JSON without manual serialization |
| **Response helpers** | `resp.IsOK()`, `resp.BodyContains("error")`, `resp.JSON(&user)` |
| **Thread-safe** | Multiple test workers can share the same client safely |
