
# qtag

`qtag` is a small, reflection-based Go package for decoding HTTP query parameters into strongly-typed structs using struct tags.

It is designed to be minimal, dependency-free, and ergonomic—similar in spirit to `encoding/json`, but for `url.Values`.

> ⚠️ **Warning**
> This package is still under active development. APIs, behavior, and tag syntax may change.

---

## Features

* Decode `url.Values` into a struct using tags
* Works directly with `*http.Request`
* Supports common Go scalar types
* Supports `encoding.TextUnmarshaler`
* Opt-in field decoding via tags
* Ignore fields explicitly
* Planned extensibility (e.g. additional tag options)

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

Given a request like:

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

Currently supported field types:

* `string`
* `bool`
* `int`
* `int64`
* `float32`
* `float64`
* Any type implementing `encoding.TextUnmarshaler`

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

### Default Values

Default values are supported via `default=...` in the tag:

```
type QueryParams struct {
    Limit int64 `qt:"limit,default=25"`
    Page  int64 `qt:"page,default=1"`
}
```

If the query parameter is missing or empty, the default value will be used.

Note: Default parsing follows the same rules as normal decoding and must be valid for the field type.

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

---
