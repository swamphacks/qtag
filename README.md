
# qtag

`qtag` is a small, reflection-based Go package for decoding HTTP query parameters into strongly-typed structs using struct tags.

It is designed to be minimal and dependency-free.

---

## Features
* Zero Dependencies: Built entirely on the Go standard library.
* Direct Integration: Seamlessly handles both *http.Request and url.Values.
* Recursive Routing: Supports deep nesting via dot-notation and flat embedding.
* Smart Fallbacks: Built-in support for default values and explicit field ignoring.
* Interface Aware: Natively supports any type implementing encoding.TextUnmarshaler, such as time.Time and net.IP.
* Strong Typing: Descriptive error reporting for type mismatches and invalid pointers.
---

## Installation

```
go get github.com/swamphacks/qtag
```

---

## Basic Usage

Add a `qt` tag to any struct field you want populated from the query string.

```
type QueryParams struct {
    Limit int64 `qt:"limit"`
    Page  int64 `qt:"page"`
}
```

Given a request:

```
https://example.com/some/endpoint?limit=200&page=2
```

You can decode it as follows:

```
var qp QueryParams
err := qtag.Decode(r, &qp)
if err != nil {
    // handle error
}
```

---

## Using `Unmarshal` Directly

If you already have `url.Values`, you can call `Unmarshal` directly:

```
values := url.Values{
    "limit": {"100"},
    "page":  {"1"},
}

var qp QueryParams
err := qtag.Unmarshal(values, &qp)
```

---

## Supported Types
We support all primitive values and anything that implements the standard `encoding.TextUnmarshaler` interface.

If a field implements `encoding.TextUnmarshaler`, it will be preferred over built-in parsing.

---

## Tag Options

### Basic Tag

```
Field string `qt:"field_name"`
```

### Ignore a Field

Use `-` to explicitly ignore a field:

```
InternalID string `qt:"-"`
```

> Note: Even if you supply a key or default, it will be ignored as long as `-` is present!!!

### Default Values

Default values are supported via `default=...` in the tag:

```
type QueryParams struct {
    Limit int64 `qt:"limit,default=25"`
    Page  int64 `qt:"page,default=1"`
}
```

If the query parameter is missing or empty, the default value will be used. If no default is provided
and the field is empty, we utilize that variable's zero value.

> Note: Default parsing follows the same rules as normal decoding and must be valid for the field type.

### Nested Structs
Since our parser is recursive, you can nest structs and parse flatly OR with a prefix.

For example:
```
type Pagination struct {
    Limit int64 `qt:"limit,default=25"`
    Page  int64 `qt:"page,default=1"`
}

type QueryParams struct {
    Paging Pagination
    UserID uuid.UUID `qt:"user_id"`
}
```

This will successfully parse `https://example.com?limit=20&page=30&user_id=dcfff211-debd-4ff6-a07d-16d29050e6ae`.

If you want namespaces, you can add a key to the struct to nest it with a namespace prefix.

For example:
```
type Options struct {
    Status *string `qt:"status,default=None"`
    Role *string `qt:"role"`
}

type QueryParams struct {
    Options Options `qt:"option"`
    UserID uuid.UUID `qt:"user_id"`
}
```

This will parse `https://example.com?option.status=ADMITTED&option.role=ADMIN&user_id=dcfff211-debd-4ff6-a07d-16d29050e6ae`

---

## Error Handling

`qtag` returns descriptive errors when:

* A non-struct pointer is provided
* A value cannot be parsed into the target type
* An unsupported field type is encountered

Example:

```
Cannot parse limit with value abc into an integer.
```
