# 📦 Funguy – Automatic multipart form decoding for Go

Funguy is a Go package that **automatically decodes** a `multipart.Form` (such as from an `http.Request`) into a **Go struct** using field tags.  
It supports primitive types, pointers, slices, time values, and file uploads.

---

## ✨ Features

- **Automatic decoding** of a multipart form into a Go struct.
- **Tag-based mapping** from form keys to struct fields.
- **Automatic type conversion** from `[]string` to Go types.
- **Full support for pointers and slices**.
- **Built-in support for `time.Time` and `*time.Time` (RFC3339 format)**.
- **File support** via `*multipart.FileHeader` and `[]*multipart.FileHeader`.

---

## 📝 Installation

```bash
go get github.com/youraccount/funguy
```
---

## 🚀 Basic Usage

1. Define your struct with `funguy` tags
```golang 
type Testing struct {
    Integer int                   `funguy:"integer"`
    Float   float32               `funguy:"float"`
    Text    string                `funguy:"text"`
    When    *time.Time            `funguy:"when"`
    Names   []string              `funguy:"names"`
    Files   []*multipart.FileHeader `funguy:"files"`
}
```
2. Use the decoder in your handler

```golang
func index(w http.ResponseWriter, r *http.Request) {
    err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var dst Testing
    decoder := funguy.NewDecoder()

    err = decoder.Decode(&dst, r.MultipartForm)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    fmt.Fprintf(w, "Decoded struct: %+v", dst)
}
```

## 📜 License
MIT License © 2025 Pangolin's_Shell