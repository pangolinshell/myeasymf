# 📦 MyEasyMultipartForm – Automatic multipart form decoding for Go

MyEasyM(ultipart)F(Form) is a Go package that **automatically decodes** a `multipart.Form` (such as from an `http.Request`) into a **Go struct** using field tags.  
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
go get github.com/pangolinshell/myeasymf
```
---

## 🚀 Basic Usage

1. Define your struct with `form` tags (you can change the tag by modifing the `Tag` variable )
```golang 
type Testing struct {
    Integer int                     `form:"integer"`
    Float   float32                 `form:"float"`
    Text    string                  `form:"text"`
    When    *time.Time              `form:"when"`
    Names   []string                `form:"names"`
    Files   []*multipart.FileHeader `form:"files"`
}
```
1. Use the decoder in your handler

```golang
func index(w http.ResponseWriter, r *http.Request) {
    err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var dst Testing
    decoder := myeasyform.NewDecoder()

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